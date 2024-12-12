// client.go
package eclipz

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"reflect"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	CONTROLLER_DEFAULT_PORT = 8443
)

var mustNotify bool = true

var done chan bool
var sendQueue chan interface{}
var interrupt chan os.Signal

var WSConn *websocket.Conn

func closeWSConn() {
	if WSConn != nil {
		WSConn.Close()
		WSConn = nil
	}
}

func WsDisconnect() {
	done <- true
}

func WsTerminate() {
	interrupt <- os.Interrupt
}

func getWSConn() *websocket.Conn {
	if WSConn == nil {
		var err error
		var dialer *websocket.Dialer = websocket.DefaultDialer
		caCert := getCAFileName()
		if caCert != "" {
			rootPEM, err := os.ReadFile(caCert)
			if err != nil {
				log.Printf("failed to read root certificate %s", caCert)
				return nil
			}
			roots := x509.NewCertPool()
			ok := roots.AppendCertsFromPEM(rootPEM)
			if !ok {
				log.Printf("failed to parse root certificate %s", caCert)
				return nil
			}
			dialer = &websocket.Dialer{TLSClientConfig: &tls.Config{RootCAs: roots}}
			log.Printf("Using Root CA from %s", caCert)
		}
		url := fmt.Sprintf("wss://%s:%d/socket", Config.Controller.Address, Config.Controller.Port)
		WSConn, _, err = dialer.Dial(url, nil)
		if err != nil {
			log.Printf("Error connecting to %s: %v\n", url, err)
			//RequestCredentials <- true
			// if mustNotify {
			// 	ErrorChannel <- "Error trying to connect to the Controller."
			// 	mustNotify = false
			// }
			return nil
		}
		log.Printf("Connected to %s", url)
	}
	return WSConn
}

func receiveHandler(conn *websocket.Conn) {
	defer close(done)
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error during message reading:", err)
			break
		}
		//log.Printf("Received: %d %s", messageType, message)
		var resp interface{}
		resp = processMessage(messageType, message)
		if isNil(resp) {
			//log.Printf("Response is nil\n")
		} else {
			bytes, err := json.Marshal(resp)
			if err == nil {
				err = conn.WriteMessage(messageType, bytes)
			}
			if err != nil {
				log.Printf("Error sending resp: %v\n", err)
			} else {
				log.Printf("Sent Response\n")
				log.Printf("%s\n", string(bytes))
			}
		}
	}
}

func getNextMsgId() string {
	return fmt.Sprintf("MSG-%d", rand.Int63())
}

func rpcSend(msg interface{}) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return
	}
	n := atomic.AddInt32(&Client.SendCountSinceLastRecv, 1)
	if n > MaxSendsSinceLastRecv {
		// Too many sends without any receives
	}
	conn := getWSConn()
	if conn == nil {
		// Unable to connect
		return
	}
	err = conn.WriteMessage(websocket.TextMessage, bytes)
	if err != nil {
		log.Println("Error during writing to websocket:", err)
		return
	}

	if reflect.TypeOf(msg) != reflect.TypeOf(RpcStats{}) {
		log.Printf("Send %d bytes: '%s'\n", len(bytes), string(bytes))
	}
}

func wsclient_start() bool {
	done = make(chan bool) // Channel to indicate that the receiverHandler is done
	sendQueue = make(chan interface{}, 100)
	interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

	signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	conn := getWSConn()
	if conn == nil {
		// Unable to connect - Try again later
		return true
	}
	defer closeWSConn()

	//loginAttempt <- false
	go receiveHandler(conn)
	go func() { time.Sleep(1 * time.Second); ClientLogin() }()

	var msg interface{}

	// Our main loop for the client
	// We send our relevant packets here
	for {
		select {
		case msg = <-sendQueue:
			// RPC to send to controller
			rpcSend(msg)

		case <-time.After(time.Duration(60) * time.Second):
			// timeout every 60 secs

		case <-done:
			log.Println("Receiver Channel Closed! Restarting....")
			return true

		case <-interrupt:
			// We received a SIGINT (Ctrl + C). Terminate gracefully...
			log.Println("Received SIGINT interrupt signal. Closing all pending connections")

			// Close our websocket connection
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("Error during closing websocket:", err)
				return false
			}

			select {
			case <-done:
				log.Println("Receiver Channel Closed! Exiting....")
			case <-time.After(time.Duration(1) * time.Second):
				log.Println("Timeout in closing receiving channel. Exiting....")
			}
			return false
		}
	}
}

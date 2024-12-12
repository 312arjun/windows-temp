package eclipz

import (
	"bytes"
	crand "crypto/rand"
	"encoding/binary"
	"log"
	mrand "math/rand"
	"time"

	"golang.zx2c4.com/wireguard/windows/tunnel"
)

const (
	InitialRetryInterval = 5 * time.Second
	MaxRetryInterval     = 300 * time.Second
)

var loginAttempt = make(chan bool)

var retryInterval time.Duration = InitialRetryInterval

func resetRetryInterval() {
	retryInterval = InitialRetryInterval
}

func EclipzInit() {
	randSeed()
	loadConfig()
	OpenLogFile()
}

func EclipzMain() {
	// Wireguard should be starting. Wait until the iface shows up.
	log.Print("EclipzMain: waiting for eclipz iface ...")
	for tmo := 10; tmo > 0; tmo++ {
		if CheckEclipzIface() {
			log.Print("EclipzMain: found eclipz iface ...")
			break
		}
		time.Sleep(time.Second)
	}

	log.Print("EclipzMain: calling WgInitialize()")
	WgInitialize(&Config.Wireguard)

	// If new private key generated, save it
	saveConfigFile()

	// go peersStatusTask()
	// go statsTask()

	Client.Status = STATUS_NOT_LOGGED_IN
	//printInterfaces()
	if Config.Webserver.Enable {
		go webServer()
	}
	for {
		loadConfig()
		if !wsclient_start() {
			// Interrupted!
			log.Print("EclipzMain: wsclient_start() failed\n")
			break
		}
		go func() {
			RequestCredentials <- true
			<-RetryControllerConnChannel
			tunnel.WGlog("Retring connection to controller...")
			loginAttempt <- true
		}()
		<-loginAttempt
	}
}

// Use crypto rand to seed math rand
func randSeed() {
	var seed int64

	r := make([]byte, 8)
	_, _ = crand.Read(r)
	buf := bytes.NewBuffer(r)
	binary.Read(buf, binary.LittleEndian, &seed)
	mrand.Seed(seed)
}

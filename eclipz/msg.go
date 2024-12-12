package eclipz

import (
	"encoding/json"
	"log"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// Return value is data to be sent to the controller in response to this message
func processTextMessage(msg []byte) interface{} {
	var common RpcCommon
	err := json.Unmarshal(msg, &common)
	if err != nil {
		log.Printf("Data is not JSON\n")
		return nil
	}
	switch common.Cmd {
	case CmdLoginResp:
		err = processLoginResp(msg)
	case CmdClientParams:
		err = processClientParams(msg)
	case CmdPeerInfo:
		err = processPeerInfo(msg)
	case CmdClientNotify:
		err = processClientNotify(msg)
	case CmdLogoutResp:
		err = processLogoutResp(msg)
	case "":
		// Empty message - Ignore
	default:
		log.Printf("Unknown Request Type %s\n", common.Cmd)
	}
	return nil
}

// BinaryMessage denotes a binary data message.
func processBinaryMessage(msg []byte) interface{} {
	return nil
}

// CloseMessage denotes a close control message. The optional message
// payload contains a numeric code and text. Use the FormatCloseMessage
// function to format a close message payload.
func processCloseMessage(msg []byte) interface{} {
	return nil
}

// PingMessage denotes a ping control message. The optional message payload
// is UTF-8 encoded text.
func processPingMessage(msg []byte) interface{} {
	return nil
}

// PongMessage denotes a pong control message. The optional message payload
// is UTF-8 encoded text.
func processPongMessage(msg []byte) interface{} {
	return nil
}

func processMessage(msgType int, msg []byte) interface{} {
	//log.Printf("Rcvd message type %d\n", msgType)
	var resp interface{}

	Client.LastRecvTime = time.Now()
	atomic.StoreInt32(&Client.SendCountSinceLastRecv, 0)
	switch msgType {
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	case websocket.TextMessage:
		resp = processTextMessage(msg)

	// BinaryMessage denotes a binary data message.
	case websocket.BinaryMessage:
		resp = processBinaryMessage(msg)

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	case websocket.CloseMessage:
		resp = processCloseMessage(msg)

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	case websocket.PingMessage:
		resp = processPingMessage(msg)

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	case websocket.PongMessage:
		resp = processPongMessage(msg)
	default:
		log.Printf("Unknown message type %d\n", msg)
	}
	return resp
}

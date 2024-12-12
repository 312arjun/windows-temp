package eclipz

import (
	"time"
	"log"
	"encoding/json"
)

const (
	// Client Status
	STATUS_NOT_LOGGED_IN = "Not Logged In"
	STATUS_LOGGED_IN = "Logged In"
	STATUS_LOGIN_SUCCESS = "Login Successful"
	STATUS_LOGIN_FAILED = "Login Failed"

	// Peer Status
	STATUS_PEER_READY = "Peer Ready"
	STATUS_PEER_ADDED = "Peer Added"
	STATUS_PEER_REMOVED = "Peer Removed"
	STATUS_PEER_CONNECTED = "Peer Connected"
	STATUS_PEER_DISCONNECTED = "Peer Disconnected"
)

const MaxSendsSinceLastRecv = 20    // Reset controller connection if too many sends without any receive

func processClientParams(msg []byte) error {
	var params RpcClientParams
	err := json.Unmarshal(msg, &params)
	if err != nil {
		log.Printf("Data is not JSON\n")
		return err
	}
	if params.Status == 0 {
		Client.ChangeStatus(STATUS_LOGGED_IN)
		Client.LoginTime = time.Now()
		Client.VirtualIP = params.VirtualIP
		Client.Apps = params.Apps
		Client.WgPublicAddress = params.WgPublicAddress
	} else {
		Client.ChangeStatus(params.StatusText)
	}
	(&Client).ConfigureInteface()

	for _, app := range params.Apps {
		sendPeerInfoReq(app)
	}
	return nil
}

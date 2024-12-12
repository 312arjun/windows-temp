package eclipz

import (
	"log"
	"encoding/json"
)

// Handle ClientReady Notify message
// Backend at this point has information regarding this clients public
// IP and UDP port used by wireguard
func processNotifyClientReady(notify *RpcClientNotify) {
	log.Printf("ClientReady Notification\n")
	ClientReadyNotification <-true
}

// Handle ServiceOnline Notify message
func processNotifyServiceOnline(notify *RpcClientNotify) {
	log.Printf("ServiceOnline Notification %s\n", notify.ServiceName)
	for _, app := range Client.Apps {
		if app.Service == notify.ServiceName {
			log.Printf("Got Notify: Send PeerInfoReq for %s\n", notify.ServiceName)
			sendPeerInfoReq(app)
			break
		}
	}
}

// Handle Notify meesage received by Client
func processClientNotify(msg []byte) error {
	var notify RpcClientNotify
	err := json.Unmarshal(msg, &notify)
	if err != nil {
		log.Printf("Data is not JSON\n")
		return err
	}
	switch notify.NotifyType {
	case NotifyServiceOnline:
		processNotifyServiceOnline(&notify)
	case NotifyClientReady:
		processNotifyClientReady(&notify)
	}
	return nil
}

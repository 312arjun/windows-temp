package eclipz

import (
	"log"
	"time"
)

const (
	// Values in database - do not change here
	DB_STATUS_ACTIVE   = "A"
	DB_STATUS_DELETED  = "D"
	DB_STATUS_DISABLED = "S"

	DB_ROLE_CLIENT     = "U"
	DB_ROLE_GATEWAY    = "S"
)

type ClientData struct {
	AuthToken    string
	Status       string
	VirtualIP    string
	LoginTime    time.Time
	LastRecvTime time.Time
	SendCountSinceLastRecv int32
	WgControllerKey string // Controller's wireguard public key
	WgPublicAddress string // Client's UDP Public IP address (as seen by controller)
	Apps         []*App
}
var Client ClientData

func GetMyDeviceId() string {
	key := WgGetMyPublicKey()
	if len(key) > 8 {
		// Truncate to 8 base64 chars
		key = key[:8]
	}
	return key
}

func (client *ClientData) ChangeStatus(status string) {
	log.Printf("Client status changed from %s to %s\n", client.Status, status)
	client.Status = status
	client.StatusChanged()
}

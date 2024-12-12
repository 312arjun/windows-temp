package eclipz

import (
	"time"
	"log"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// We configure a wireguard instance from client to the controller. The wireguard tunnel is initiated to
// the backend, but the tunnel is not actually established. controller only receives the initial wireguard
// message and never responds to any wireguard messages.
// This is used to determine the public IP and externl UDP port that wireguard uses
// publicKey is controller's wireguard public key
func configureBackendWireguard(publicKey string) *Peer {
	peer := Peer{}
	peer.Name = "controller"
	peer.Status = STATUS_PEER_READY
	peer.Role = RoleService
	peer.Initiate = true
	peer.WgKey = publicKey
	peer.PublicIP = Config.Controller.Address
	peer.LocalIPs = nil
	peer.AllowedIPs = nil
	peer.StartTime = time.Now()

	// Determine Endpoint based on peer's addresses
	peer.SetEndpoint()

	// Configure Wireguard
	WgConfigurePeer(&peer)
	return &peer
}

func genRandomPublicKey() string {
	privkey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Fatalf("Failed to generate private key: %v\n", err)
	}

	pubkey := privkey.PublicKey()
	return pubkey.String()
}

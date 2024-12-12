package eclipz

import (
	"fmt"
	"log"
	"sync"
	"time"
	"encoding/json"
)
const (
	WG_DEFAULT_UDP_PORT = 51820
)

type Peer struct {
	Name       string
	Role       string
	EnclaveId  string
	Initiate   bool
	Status     string
	WgDeviceName string
	WgKey      string
	WgPrevKey  string
	WgPresharedKey string
	WgIPAddress string
	WgPort     int
	Endpoint   string
	VirtualIP  string
	PublicIP   string
	LocalIPs   []string
	Apps       []*App
	AllowedIPs []string
	StartTime  time.Time
	LastHandshakeTime  time.Time
	PrevRxBytes int64
	PrevTxBytes int64
	RxBytes    int64
	TxBytes    int64
}

var Peers map[string]*Peer = make(map[string]*Peer)
var PeersMutex sync.RWMutex

func peerNew(name string) *Peer {
	peer := &Peer{}
	peer.Name = name

	PeersMutex.Lock()
	Peers[peer.Name] = peer
	PeersMutex.Unlock()
	return peer
}

func peerGet(name string) *Peer {
	PeersMutex.Lock()
	peer := Peers[name]
	PeersMutex.Unlock()
	return peer
}

func sendPeerInfoReq(app *App) {
	req := &RpcPeerInfoReq{}
	req.Cmd = CmdPeerInfoReq
	req.MsgId = getNextMsgId()
	req.Token = Client.AuthToken
	req.App = *app

	sendQueue<-req
}

// Handle PeerInfo meesage received by either Client or Service
func processPeerInfo(msg []byte) error {
	log.Printf("RCVD: %s\n", string(msg))
	var peerInfo RpcPeerInfo
	err := json.Unmarshal(msg, &peerInfo)
	if err != nil {
		log.Printf("PeerInfo: Data is not JSON\n")
		return err
	}

	peer := peerGet(peerInfo.PeerName)
	if peer == nil {
		peer = peerNew(peerInfo.PeerName)
	} else {
		// Save previous public key to remove from wireguard
		peer.WgPrevKey = peer.WgKey
	}
	if peerInfo.Status != 0 {
		peer.ChangeStatus(peerInfo.StatusText)
	} else {
		peer.ChangeStatus(STATUS_PEER_READY)
		peer.Role = peerInfo.Role
		peer.EnclaveId = peerInfo.EnclaveId
		peer.Initiate = peerInfo.Initiate
		peer.WgKey = peerInfo.WgKey
		peer.WgPresharedKey = peerInfo.WgPresharedKey
		peer.WgIPAddress = peerInfo.WgIPAddress
		peer.WgPort = peerInfo.WgPort
		peer.VirtualIP = peerInfo.VirtualIP
		peer.PublicIP = peerInfo.PublicIP
		peer.LocalIPs = peerInfo.LocalIPs
		peer.AllowedIPs = peerInfo.AllowedIPs

		// Determine Endpoint based on peer's addresses
		peer.SetEndpoint()

		// Set up routes based on allowed IPs
		peer.addRoutes()

		// Configure Wireguard
		WgConfigurePeer(peer)
		peer.StartTime = time.Now()
	}


	return nil
}

func (peer *Peer) SetEndpoint() {
	addr := peer.WgIPAddress
	if addr == "" {
		addr = peer.PublicIP
	}
	port := peer.WgPort
	if port == 0 {
		port = WG_DEFAULT_UDP_PORT
	}
	peer.Endpoint = fmt.Sprintf("%s:%d", addr, port)
}

func (peer *Peer) ChangeStatus(status string) {
	log.Printf("Peer %s status changed from %s to %s\n", peer.Name, peer.Status, status)
	peer.Status = status
	peer.StatusChanged()
}

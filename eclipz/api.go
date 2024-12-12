package eclipz

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type ApiStatus struct {
	Name              string `json:"name,omitempty"`
	Domain            string `json:"domain,omitempty"`
	Role              string `json:"role,omitempty"`
	State             string `json:"state,omitempty"`
	VirtualIP         string `json:"virtual_ip,omitempty"`
	Device            string `json:"device,omitempty"`
	ControllerAddress string `json:"controller_address,omitempty"`
	ControllerPort    int    `json:"controller_port,omitempty"`
	PublicAddress     string `json:"public_address,omitempty"`
	PublicKey         string `json:"public_key,omitempty"`
	TimeSinceLogin    string `json:"time_since_login,omitempty"`
	TimeSinceLastRecv string `json:"time_since_lastrecv,omitempty"`
	Apps              []*App `json:"apps,omitempty"`
}

type ApiPeer struct {
	Name              string `json:"name,omitempty"`
	Endpoint          string `json:"endpoint,omitempty"`
	VirtualIP         string `json:"virtual_ip,omitempty"`
	LocalIPs          string `json:"local_ips,omitempty"`
	Uptime            string `json:"uptime,omitempty"`
	WgKey             string `json:"public_key,omitempty"`
	AllowedIPs        string `json:"allowed_ips,omitempty"`
	WgIPAddress       string `json:"wg_ip_addr,omitempty"`
	WgPort            int    `json:"wg_port,omitempty"`
	RxBytes           int64  `json:"rx_bytes,omitempty"`
	TxBytes           int64  `json:"tx_bytes,omitempty"`
	LastHandshakeSecs int    `json:"last_handshake_secs,omitempty"`
	Status            string `json:"status,omitempty"`
}

// Send notifications to the application via channels
var ClientStatusNotification chan *ApiStatus = make(chan *ApiStatus)
var PeerStatusNotification chan *ApiPeer = make(chan *ApiPeer)

func timeSince(t time.Time) string {
	var secs, mins, hrs int
	if t.IsZero() {
		return "??:??:??"
	}
	secs = int(time.Since(t).Seconds())
	if secs >= 60 {
		mins = secs / 60
		secs = secs % 60
	}
	if mins >= 60 {
		hrs = mins / 60
		mins = mins % 60
	}
	return fmt.Sprintf("%d:%02d:%02d", hrs, mins, secs)
}

func GetStatus() *ApiStatus {
	status := &ApiStatus{}
	status.Name = Config.Client.Name
	status.Domain = Config.Client.Domain
	if Config.Client.Role == DB_ROLE_CLIENT {
		status.Role = "Client"
	} else if Config.Client.Role == DB_ROLE_GATEWAY {
		status.Role = "Gateway"
	} else {
		status.Role = Config.Client.Role
	}
	status.State = Client.Status
	status.VirtualIP = Client.VirtualIP
	status.ControllerAddress = Config.Controller.Address
	status.ControllerPort = Config.Controller.Port
	status.Device = Config.Wireguard.Device
	if Client.WgPublicAddress != "" {
		status.PublicAddress = Client.WgPublicAddress
	} else {
		status.PublicAddress = fmt.Sprintf("%s:%d", myPublicIP, WgListenPort)
	}
	status.PublicKey = WgGetMyPublicKey()
	status.TimeSinceLogin = timeSince(Client.LoginTime)
	status.TimeSinceLastRecv = timeSince(Client.LastRecvTime)

	for _, app := range Client.Apps {
		a := &App{}
		a.Name = app.Name
		a.Service = app.Service
		a.AllowedIPs = app.AllowedIPs
		status.Apps = append(status.Apps, a)
	}
	return status
}

func GetPeers() []*ApiPeer {
	var peers []*ApiPeer

	wgPeers := WgGetPeers()

	PeersMutex.Lock()
	for _, peer := range Peers {
		if peer.Name == "" {
			continue
		}
		p := &ApiPeer{}
		p.Name = peer.Name
		p.Endpoint = peer.Endpoint
		p.VirtualIP = peer.VirtualIP
		p.Status = peer.Status
		p.Uptime = timeSince(peer.StartTime)

		// Local IPs
		var ip2 string
		for _, ip := range peer.LocalIPs {
			if ip2 != "" {
				ip2 = ip2 + ", "
			}
			ip2 = ip2 + ip
		}
		p.LocalIPs = ip2
		p.WgKey = peer.WgKey

		// Allowed IPs
		ip2 = ""
		for _, ip := range peer.AllowedIPs {
			if ip2 != "" {
				ip2 = ip2 + ", "
			}
			ip2 = ip2 + ip
		}
		p.AllowedIPs = ip2
		wgPeer := wgPeers[peer.WgKey]
		if wgPeer != nil {
			p.RxBytes = wgPeer.RxBytes
			p.TxBytes = wgPeer.TxBytes
			p.LastHandshakeSecs = int(time.Since(wgPeer.LastHandshakeTime).Seconds())
		}
		peers = append(peers, p)
	}
	PeersMutex.Unlock()
	return peers
}

func GetDeviceName() string {
	return getWgDeviceName()
}

func GetStatusJSON() string {
	var bytes []byte

	status := GetStatus()
	if status == nil {
		return ""
	}
	bytes, _ = json.MarshalIndent(status, "", "    ")
	return string(bytes)
}

func GetPeersJSON() string {
	var bytes []byte

	peers := GetPeers()
	if len(peers) == 0 {
		return "[]"
	}
	bytes, _ = json.MarshalIndent(peers, "", "    ")
	return string(bytes)
}

// Connect to contoller, login and get avaliable apps
func Connect() error {
	log.Printf("Connect Called\n")
	return nil
}

// Disconnect from contoller (tunnels stay up)
func Disconnect() error {
	log.Printf("Disconnect Called\n")
	WsDisconnect()
	return nil
}

// Disconnect from contoller and terminate application
func Terminate() error {
	log.Printf("Terminate Called\n")
	return nil
}

// Create tunnel to peer
func PeerConnect(peerName string) error {
	log.Printf("PeerConnect %s Called\n", peerName)
	return nil
}

// Disconnect tunnel to peer
func PeerDisconnect(peerName string) error {
	log.Printf("PeerDisconnect %s Called\n", peerName)
	return nil
}

// Disconnect tunnel to all peers
func PeerDisconnectAll() error {
	log.Printf("PeerDisconnectAll Called\n")
	return nil
}

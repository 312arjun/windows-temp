package eclipz

import (
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

var WgDevices map[string]*wgtypes.Device = make(map[string]*wgtypes.Device)
var WgMutex sync.RWMutex

// var zeroKey [wgtypes.KeyLen]byte
var wgClient *wgctrl.Client
var WgListenPort int
var defaultKeepalive time.Duration = 30 * time.Second

// Returns true if a new private key generated, else returns false
func WgInitialize(wgConfig *WgConfig) {
	EClog("WgInitialize(), locking mutex")
	WgMutex.Lock()
	device := WgDevices[wgConfig.Device]
	if device == nil {
		device = &wgtypes.Device{}
		device.Name = wgConfig.Device
		WgDevices[device.Name] = device
	}
	device.ListenPort = wgConfig.Port
	WgMutex.Unlock()
	EClog("WgInitialize(), unlocked mutex, private key: %s", wgConfig.PrivateKey)

	var privkey wgtypes.Key
	var err error
	if wgConfig.PrivateKey == "" {
		// No private key configured, generate new private key
		privkey, err = wgtypes.GeneratePrivateKey()
		if err != nil {
			log.Fatalf("Failed to generate private key: %v\n", err)
		}
		configChanged()
		wgConfig.PrivateKey = privkey.String()
		EClog("Device %s Generated PrivateKey %s\n", device.Name, wgConfig.PrivateKey)
	} else {
		// private key configured
		privkey, err = wgtypes.ParseKey(wgConfig.PrivateKey)
		if err != nil {
			log.Fatalf("Failed to parse private key '%s' : %v\n", wgConfig.PrivateKey, err)
		}
	}

	pubkey := privkey.PublicKey()
	device.PrivateKey = privkey
	device.PublicKey = pubkey

	wgConfig.PublicKey = pubkey.String()
	EClog("Device %s PublicKey %s\n", device.Name, wgConfig.PublicKey)

	cfg := wgtypes.Config{
		PrivateKey:   &device.PrivateKey,
		ReplacePeers: true,
	}
	if wgConfig.Port != 0 {
		cfg.ListenPort = &wgConfig.Port
	}

	if wgConfig.FWmark != 0 {
		cfg.FirewallMark = &wgConfig.FWmark
	}

	err = WgConfigure(device.Name, cfg)
	if err != nil {
		EClog("Failed to configure device %s: %v", device.Name, err)
		return
	}
	EClog("Configured wireguard device %s public-key %s\n", device.Name, wgConfig.PublicKey)
	WgPrintAll()
}

func WgGetMyPublicKey() string {
	device := getDevice()
	if device == nil {
		return ""
	}
	return device.PublicKey.String()
}

func getDevice() *wgtypes.Device {
	WgMutex.Lock()
	defer WgMutex.Unlock()
	for _, device := range WgDevices {
		return device
	}
	return nil
}

func getDeviceName() string {
	device := getDevice()
	if device == nil {
		return ""
	}
	return device.Name
}

func getClient() (*wgctrl.Client, error) {
	if wgClient == nil {
		c, err := wgctrl.New()
		if err != nil {
			return nil, err
		}
		wgClient = c
		EClog("Connected to Wireguard\n")
	}
	return wgClient, nil
}

func closeClient() {
	if wgClient != nil {
		wgClient.Close()
		wgClient = nil
		EClog("Disconnected from Wireguard\n")
	}
}

func WgConfigurePeer(peer *Peer) error {
	device := getDevice()
	return AddPeer(device, peer)
}

// Add a Peer config to Wireguard
func AddPeer(device *wgtypes.Device, peer *Peer) error {
	var psk *wgtypes.Key
	var wgk wgtypes.Key

	// First remove existing configuration for this peer
	_ = RemovePeer(device, peer)

	var keepalive time.Duration
	if Config.Wireguard.Keepalive == 0 {
		keepalive = defaultKeepalive
	} else {
		keepalive = time.Duration(Config.Wireguard.Keepalive) * time.Second
	}

	// Convert from strings to internal type
	wgk, err := wgtypes.ParseKey(peer.WgKey)
	if err != nil {
		EClog("ERROR base64 decode of WgKey '%s'\n", peer.WgKey)
		return err
	}

	if peer.WgPresharedKey != "" {
		k2, err := wgtypes.ParseKey(peer.WgPresharedKey)
		if err != nil {
			EClog("ERROR base64 decode of PresharedKey '%s'\n", peer.WgPresharedKey)
		}
		psk = &k2
	}

	endpointUDP, err := net.ResolveUDPAddr("udp", peer.Endpoint)
	if err != nil {
		EClog("ERROR resolving UDP address '%s'\n", peer.Endpoint)
		return err
	}

	var allowedIPs []net.IPNet
	if peer.VirtualIP == "" {
		var ipNet net.IPNet
		ipNet.IP = endpointUDP.IP
		ipNet.Mask = net.IPv4Mask(255, 255, 255, 255)
		allowedIPs = append(allowedIPs, ipNet)
	} else {
		_, ipNet, err := net.ParseCIDR(peer.VirtualIP + "/32")
		if err != nil {
			EClog("ERROR parsing peer's VirtualIP address '%s'\n", peer.VirtualIP)
		} else {
			allowedIPs = append(allowedIPs, *ipNet)
		}
	}

	for _, ips := range peer.AllowedIPs {
		if ips == "" {
			// Ignore empty strings
			continue
		}
		addrs := strings.Split(ips, ";")
		for _, a := range addrs {
			// Trim white space from start and end of address
			ipStr := strings.Trim(a, " ")
			if !strings.Contains(ipStr, "/") {
				ipStr = ipStr + "/32"
			}
			lowIndex := strings.Index(ipStr, "-")      //Index to remove the protocol
			highIndex := strings.LastIndex(ipStr, ":") //index prev to the port info
			ipStr = ipStr[lowIndex+1 : highIndex]
			_, ipNet, err := net.ParseCIDR(ipStr)
			if err != nil {
				EClog("ERROR parsing AllowedIP address '%s'\n", ipStr)
				continue
			}
			allowedIPs = append(allowedIPs, *ipNet)
		}
	}

	pcfg := wgtypes.PeerConfig{
		PublicKey:                   wgk,
		ReplaceAllowedIPs:           true,
		PresharedKey:                psk,
		PersistentKeepaliveInterval: &keepalive,
		AllowedIPs:                  allowedIPs,
	}
	if peer.Initiate {
		// Set endpoint only if initiating
		pcfg.Endpoint = endpointUDP
	} else {
		pcfg.Endpoint = endpointUDP
	}
	cfg := wgtypes.Config{
		ReplacePeers: false,
		Peers:        []wgtypes.PeerConfig{pcfg},
	}
	err = WgConfigure(device.Name, cfg)
	if err != nil {
		EClog("failed to configure Peer %s: %v", peer.Name, err)
		return err
	}
	EClog("Added peer %s %s %s (%s) I=%t %s to wireguard\n", peer.Name,
		peer.PublicIP, peer.Endpoint, endpointUDP.String(), peer.Initiate, peer.WgKey)
	return nil
}

func WgRemovePeer(peer *Peer) error {
	device := getDevice()
	return RemovePeer(device, peer)
}

// Remove  Peer config from Wireguard
func RemovePeer(device *wgtypes.Device, peer *Peer) error {
	pkey, err := wgtypes.ParseKey(peer.WgKey)
	if err != nil {
		EClog("Bad Public Key '%s' for peer %s\n", peer.WgKey, peer.Name)
		return err
	}
	cfg := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{{
			Remove:    true,
			PublicKey: pkey,
		}},
	}
	err = WgConfigure(device.Name, cfg)
	if err != nil {
		EClog("failed to remove peer %s: %v", peer.Name, err)
		closeClient()
		return err
	}
	EClog("Removed peer %s %s from wireguard\n", peer.Name, peer.WgKey)
	return nil
}

func WgConfigure(deviceName string, cfg wgtypes.Config) error {
	c, err := getClient()
	if err != nil {
		return err
	}
	err = c.ConfigureDevice(deviceName, cfg)
	if err != nil {
		EClog("failed to configure device %s: %v", deviceName, err)
		closeClient()
		return err
	}
	return nil
}

func WgPrintAll() {
	c, err := getClient()
	if err != nil {
		return
	}
	var devices []*wgtypes.Device
	devices, err = c.Devices()
	if err != nil {
		EClog("failed to get devices: %v", err)
		closeClient()
		return
	}

	for _, d := range devices {
		printDevice(d)
		if WgListenPort == 0 {
			WgListenPort = d.ListenPort
		}

		for _, p := range d.Peers {
			printPeer(&p)
		}
	}
}

func printDevice(d *wgtypes.Device) {
	EClog("interface: %s (%s) port %d %s\n",
		d.Name,
		d.Type.String(),
		d.ListenPort,
		d.PublicKey.String())
}

func printPeer(p *wgtypes.Peer) {
	EClog("   peer: %s endpoint: %s allowed ips: %s latest-handshake: %s transfer: %d B received, %d B sent\n",
		p.PublicKey.String(),
		p.Endpoint.String(),
		ipsString(p.AllowedIPs),
		p.LastHandshakeTime.String(),
		p.ReceiveBytes,
		p.TransmitBytes)
}

func ipsString(ipns []net.IPNet) string {
	ss := make([]string, 0, len(ipns))
	for _, ipn := range ipns {
		ss = append(ss, ipn.String())
	}

	return strings.Join(ss, ", ")
}

type WgDevice struct {
	Name      string
	Port      int
	PublicKey string
}

type WgPeer struct {
	DeviceName        string
	PublicKey         string
	Endpoint          string
	AllowedIPs        string
	RxBytes           int64
	TxBytes           int64
	LastHandshakeTime time.Time
}

func WgGetDevices() []*WgDevice {
	c, err := getClient()
	if err != nil {
		return nil
	}
	var devices []*wgtypes.Device
	devices, err = c.Devices()
	if err != nil {
		EClog("failed to get devices: %v", err)
		closeClient()
		return nil
	}

	var wgDevices []*WgDevice

	for _, d := range devices {
		WgMutex.RLock()
		_, ok := WgDevices[d.Name]
		WgMutex.RUnlock()
		if !ok {
			// Ignore this wireguard device, it is not for this instance of client
			continue
		}

		wd := WgDevice{}
		wd.Name = d.Name
		wd.Port = d.ListenPort
		wd.PublicKey = d.PublicKey.String()
		wgDevices = append(wgDevices, &wd)
	}
	return wgDevices
}

func WgGetPeers() map[string]*WgPeer {
	var wgPeers map[string]*WgPeer = make(map[string]*WgPeer)

	c, err := getClient()
	if err != nil {
		return wgPeers
	}
	var devices []*wgtypes.Device
	devices, err = c.Devices()
	if err != nil {
		EClog("failed to get devices: %v", err)
		closeClient()
		return wgPeers
	}

	for _, d := range devices {
		WgMutex.RLock()
		_, ok := WgDevices[d.Name]
		WgMutex.RUnlock()
		if !ok {
			// Ignore this wireguard device, it is not for this instance of client
			continue
		}
		for _, p := range d.Peers {
			wp := WgPeer{}
			wp.DeviceName = d.Name
			wp.PublicKey = p.PublicKey.String()
			wp.Endpoint = p.Endpoint.String()
			wp.AllowedIPs = ipsString(p.AllowedIPs)
			wp.LastHandshakeTime = p.LastHandshakeTime
			wp.RxBytes = p.ReceiveBytes
			wp.TxBytes = p.TransmitBytes
			wgPeers[wp.PublicKey] = &wp
		}
	}
	return wgPeers
}

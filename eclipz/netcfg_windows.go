package eclipz

import (
	"bytes"
	"fmt"
	"log"
	"net/netip"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	w "golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

/* This file is accessed from eclipzd code via these functions:
- func myLocalAddresses() []string {
- func (client *ClientData) ConfigureInteface()
- func (peer *Peer) addRoutes()
- func getConfigFileName() string
- func getLogFileName() string
- func getWgDeviceName() string
- func getCAFileName() string
- func CheckEclipzIface() - see if the eclipz0 iface exists.


  The other functions herein are local helpers.
*/

var homeDir string = "C:\\Program Files\\Eclipz"

func getConfigFileName() string {
	filename := fmt.Sprintf("%s\\config.json", homeDir)
	log.Printf("getConfigFileName: %s", filename)
	return filename
}

func getLogFileName() string {
	filename := fmt.Sprintf("%s\\eclipz.log", homeDir)
	log.Printf("getLogFileName: %s", filename)
	return filename
}

func getCAFileName() string {
	certpath := Config.Controller.CACert

	// get rid of prepended "/etc/""
	names := strings.Split(certpath, "/")
	certfile := names[len(names)-1]
	ca := fmt.Sprintf("%s\\%s", homeDir, certfile)
	log.Printf("getCAFileName: %s", ca)
	return ca
}

func getWgDeviceName() string {
	log.Printf("getWgDeviceName: %s", Config.Wireguard.Device)
	return Config.Wireguard.Device
}

/* Entry: Get IP addresses of local (ethernet) interface.

	From Saroop:
The idea is to include all interfaces by which someone can connect from outside
the system. So include both IPv4 and IPv6 addresses for all such interfaces.
Its not that critical at this point. This is mainly for future use

	Sample addresses return from netcfg_linux log:
netcfg_linux.go:163: Addresses: [172.31.1.108/20 fe80::873:c4ff:fe10:3173/64]

	On Error, just log the error and return nil
*/

type primaryAdapterInfo struct {
	ipAddress  string
	subnetMask string
	gateway    string
}

var primaryAdapter primaryAdapterInfo

func myLocalAddresses() []string {
	var addresses []string // return value

	// Get adapter list with general info
	ifaces, err := getIfacesInfo()
	if err != nil {
		log.Printf("getIfacesInfo; error %s", err)
		log.Fatal(err)
	}

	// Get adapter list with IPv6 address info.
	Addresses, err := getIfaceAddresses()
	if err != nil {
		log.Fatal(err)
		log.Printf("getIfacesInfo; error %s", err)
	}

	// loop through ifaces list
	for _, iface := range ifaces {

		// Skip any that are not WiFi or Ethernet.
		if (iface.Type != 6) && (iface.Type != 71) {
			continue
		}
		addressList := iface.IpAddressList
		gateways := iface.GatewayList

		// Interfaces without IP address set are skipped
		ipAddress := bp2print(addressList.IpAddress.String[:])
		if ipAddress == string("0.0.0.0") {
			continue
		}

		// If gateway is not set, this is not a good interface to report
		// If indicates VMware or similar things
		gateway := bp2print(gateways.IpAddress.String[:])
		if gateway == string("0.0.0.0") {
			continue
		}

		description := bp2print(iface.Description[:])
		EClog("myLocalAddresses: found Interface: %s, IpAddress (v4): %s, Gateway: %s",
			description, ipAddress, gateway)

		// Prepare the address Set for this adapter. Start with IPv4 address
		addressSet := ipAddress
		subnetMask := bp2print(addressList.IpMask.String[:])
		subnetCt := mask2snbits(subnetMask)
		addressSet += "/"
		addressSet += strconv.Itoa((int)(subnetCt))

		// Save this interface for adjustments to routing table later
		if primaryAdapter.ipAddress == "" {
			primaryAdapter = primaryAdapterInfo{ipAddress, subnetMask, gateway}
			EClog("myLocalAddresses: set primaryAdapter %+v", primaryAdapter)
		}

		/* Find the Address structure for this interface - it has stuff that is missing
		 * in the getIfaceInfo return, like IPv6 addresses.
		 */
		var addrdata *windows.IpAdapterAddresses
		for _, addrdata = range Addresses {
			if addrdata.IfIndex == iface.Index {

				// found the adapters info, now look for an IPv6 address
				for thisAddr := addrdata.FirstUnicastAddress; thisAddr != nil; thisAddr = thisAddr.Next {
					sa := thisAddr.Address.Sockaddr
					if sa.Addr.Family == windows.AF_INET6 {
						addr := thisAddr.Address.IP()
						addressSet += " "
						addressSet += addr.String()
					}
				}
				break // Done with Adapter list
			}
		}

		addresses = append(addresses, addressSet)
	}
	log.Printf("addresses: %+v", addresses)

	return addresses
}

// Convert Subnet mask, EG "255.255.128.0" to number of bits.
func mask2snbits(snmask string) int {

	nums := strings.Split(snmask, ".")
	// We expect 4 octets in text form
	if len(nums) != 4 {
		log.Print("mask2snbits: bad input", snmask)
		return 0
	}
	snbits := 0
	for i := 0; i < 4; i++ {
		bmask, _ := strconv.Atoi(nums[i])
		for bmask > 0 {
			bmask >>= 1
			snbits += 1
		}

	}

	return snbits
}

// convert a number of bits to text subnet mask, eg 17 -> "255.255.128.0"
func snbits2mask(bitCt int) string {

	mask := ""
	for byteCt := 0; byteCt < 4; byteCt++ {
		if bitCt >= 8 {
			bitCt -= 8
			mask += "255"
		} else {
			bitMask := 0
			for bitCt > 0 { // make mask byte from bit count, eg 2 -> 11000000
				bitMask >>= 1
				bitMask |= 0x80
				bitCt--
			}
			mask += strconv.Itoa(bitMask)
		}
		if byteCt < 3 {
			mask += "."
		}
	}

	return mask
}

// Convert the text-containing byte arrays we get from Windows for printing.
func bp2print(bp []byte) string {
	dlen := bytes.Index(bp, make([]byte, 1))
	return string(bp[:dlen])
}

func getIfacesInfo() ([]*windows.IpAdapterInfo, error) {
	var buf []byte
	len := uint32(15000) // recommended initial size

	buf = make([]byte, len)
	err := windows.GetAdaptersInfo((*windows.IpAdapterInfo)(unsafe.Pointer(&buf[0])), &len)
	if err != nil {
		return nil, os.NewSyscallError("getIfacesInfo", err)
	}
	if len == 0 {
		log.Printf("getIfacesInfo: no ifaces found \n")
		return nil, os.NewSyscallError("getIfacesInfo", err)
	}
	var aas []*windows.IpAdapterInfo
	for aa := (*windows.IpAdapterInfo)(unsafe.Pointer(&buf[0])); aa != nil; aa = aa.Next {
		aas = append(aas, aa)
	}
	return aas, nil
}

func getIfaceAddresses() ([]*windows.IpAdapterAddresses, error) {
	var b []byte
	l := uint32(15000) // recommended initial size
	for {
		b = make([]byte, l)
		err := windows.GetAdaptersAddresses(syscall.AF_UNSPEC, windows.GAA_FLAG_INCLUDE_PREFIX, 0,
			(*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])), &l)
		if err == nil {
			if l == 0 {
				return nil, nil
			}
			break
		}
		if err.(syscall.Errno) != syscall.ERROR_BUFFER_OVERFLOW {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
		if l <= uint32(len(b)) {
			return nil, os.NewSyscallError("getadaptersaddresses", err)
		}
	}
	var aas []*windows.IpAdapterAddresses
	for aa := (*windows.IpAdapterAddresses)(unsafe.Pointer(&b[0])); aa != nil; aa = aa.Next {
		aas = append(aas, aa)
	}
	return aas, nil
}

func bytePtrToString(p *uint8) string {
	a := (*[10000]uint8)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(a[:i])
}

// Entry: Set Virtual IP address and related network configuration for VPN interface

func (client *ClientData) ConfigureInteface() {

	EClog("Windows ConfigureInteface; ClientData: %+v\n", client)

	ifName := getWgDeviceName()

	if !CheckEclipzIface() {
		EClog("Windows ConfigureInteface: no interface %s\n", ifName)
		return // error ?
	}

	ipSpecs := strings.Split(client.VirtualIP, "/")
	ipAddr := ipSpecs[0]
	snBits, _ := strconv.Atoi(ipSpecs[1])
	ipMask := snbits2mask(snBits)
	ipGateway := "0.0.0.0" // Dont know gateway yet

	setIpApp := "netsh"
	setIpArgs := fmt.Sprintf("interface ip set address name=\"%s\" static %s %s %s",
		ifName, ipAddr, ipMask, ipGateway)

	EClog("%s %s\n", setIpApp, setIpArgs)

	args := strings.Split(setIpArgs, " ")
	cmd := exec.Command(setIpApp, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8])
	stdout, err := cmd.Output()

	if err != nil {
		EClog("ConfigureInteface; error: %s\n", err)
		EClog("ConfigureInteface; stdout: %s\n", stdout)
		log.Fatal(err)
	}
	EClog("ConfigureInteface, done. stdout: %s\n", stdout)

}

// Entry: check for existence of eclipz iface (eclipz0)
func CheckEclipzIface() bool {

	ifName := getWgDeviceName()
	log.Printf("Windows CheckEclipzIface: %s \n", ifName)

	getIfaceApp := "netsh"
	getIfaceArgs := fmt.Sprintf("interface show interface name=\"%s\"", ifName)
	args := strings.Split(getIfaceArgs, " ")
	cmd := exec.Command(getIfaceApp, args[0], args[1], args[2], args[3])
	stdout, err := cmd.Output()
	if err != nil {
		return false
	}
	//fmt.Printf("CheckEclipzIface: %s\n", stdout)
	return strings.Contains(string(stdout), ifName)
}

// Entry: Add routes

func (peer *Peer) addRoutes() {

	// First set up routes so non-VIP net addressses go to regular non-tunnel interface
	// Jira ticket E2-250
	if primaryAdapter.ipAddress != "" {
		EClog("addRoutes - adding non-VIP fixs ")

		// This uses shell calls, because it's simple and easy to test. When we have
		// time we may want to switch over to an API based solution.

		//routeCmd := "route delete 0.0.0.0"	// remove the entry wireguard just added
		cmd := exec.Command("route", "delete", "0.0.0.0")
		stdout, err := cmd.Output()
		if err != nil {
			EClog("addRoutes add error: %s", err, stdout)
		}

		cmd = exec.Command("route", "add", "0.0.0.0", "MASK", "0.0.0.0",
			primaryAdapter.gateway, "METRIC", "2")

		stdout, err = cmd.Output()
		if err != nil {
			EClog("addRoutes add error: %s", err, stdout)
		}
	}

	EClog("Windows addRoutes; Peer: %+v\n", peer)
	deviceName := getDeviceName()
	gateway := peer.VirtualIP
	for _, subnet := range peer.AllowedIPs {
		if subnet == "" {
			continue
		}
		for _, ip := range strings.Split(subnet, ";") {
			if subnet == gateway {
				continue
			}
			//Given the change in the format of the policy, I have to parse
			//the string to get only the subnet with the mask
			lowIndex := strings.Index(ip, "-")      //Index to remove the protocol
			highIndex := strings.LastIndex(ip, ":") //index prev to the port info
			ip = ip[lowIndex+1 : highIndex]
			addRoute(deviceName, ip, gateway)
		}
	}
}

func addRoute(interfaceName string, dest_network string, nextHop string) error {

	log.Printf("addRoute: iface: %s, dest: %s, next hop%s",
		interfaceName, dest_network, nextHop)

	var err error
	deviceName := getWgDeviceName()

	// Add route
	route := &w.RouteData{}
	route.Metric = 0
	route.Destination, err = netip.ParsePrefix(dest_network)
	if err != nil {
		return err
	}
	route.NextHop, err = netip.ParseAddr(nextHop)
	if err != nil {
		return err
	}

	// Get interface by name
	ifc, err := getInterfaceByName(deviceName)
	if err != nil {
		return err
	}

	// Try to route this
	_, err = ifc.LUID.Route(route.Destination, route.NextHop)
	if err == nil {
		return fmt.Errorf("LUID.Route() returned a route although it isn't added yet. Have you set route appropriately?")
	} else if err != windows.ERROR_NOT_FOUND {
		return fmt.Errorf("LUID.Route() returned an error: %v", err)
	}

	// Register a callback for getting the result
	created := make(chan bool)
	cb, err := w.RegisterRouteChangeCallback(func(notificationType w.MibNotificationType, route *w.MibIPforwardRow2) {
		switch notificationType {
		case w.MibAddInstance:
			created <- true
		}
	})
	if err != nil {
		return err
	}
	defer cb.Unregister()

	// Add the route
	err = ifc.LUID.AddRoute(route.Destination, route.NextHop, route.Metric)
	if err != nil {
		return err
	}

	select {
	case <-created:
	case <-time.After(500 * time.Millisecond):
		return fmt.Errorf("timeout waiting for route to be created")
	}

	return nil

	// Format of windows route command
	// route add 100.64.0.0 MASK 255.255.0.0 100.64.0.1 METRIC 3

	// Fill this in when we get some allowedIPs
}

// Get interace by the friendly name
func getInterfaceByName(deviceName string) (*w.IPAdapterAddresses, error) {
	ifcs, err := w.GetAdaptersAddresses(windows.AF_UNSPEC, w.GAAFlagIncludeAll)
	if err != nil {
		return nil, err
	}

	marker := strings.ToLower(deviceName)
	for _, ifc := range ifcs {
		if strings.Contains(strings.ToLower(ifc.FriendlyName()), marker) {
			return ifc, nil
		}
	}

	return nil, windows.ERROR_NOT_FOUND
}

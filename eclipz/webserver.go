package eclipz

import (
	"fmt"
	"time"
	"log"
	"strings"
	"encoding/json"
	"net/http"
)

// Web Server for Administration of Client
var headerHtml =`
<!DOCTYPE html>
<html>
<head>
<style>
.button {
  border: none;
  color: white;
  padding: 15px 32px;
  text-align: center;
  text-decoration: none;
  display: inline-block;
  font-size: 16px;
  margin: 4px 2px;
  cursor: pointer;
}

.button1 {background-color: #4CAF50;} /* Green */
.button2 {background-color: #008CBA;} /* Blue */
</style>
<title>Eclipz Client</title>
</head>
<body style="background-color:powderblue;">

<form action="http://%s/status">
    <input class="button button2" type="submit" value="Status   " />
</form>

<form action="http://%s/peers">
    <input class="button button2" type="submit" value="Peers    " />
</form>
`
var footerHtml =`
</body>
</html>
`
var tableStyle string = `<style>table, th, td {
  border: 1px solid black;
}
th, td {
  padding: 10px;
}
tr:hover {background-color: #D6EEEE;}
</style>`
// <input class="button button2" type="submit" value="peers  " />

var listenAddress string

func apiShowPeers(w http.ResponseWriter, r *http.Request) {
	var peers []*ApiPeer

	peers = GetPeers()
	httpSendResponse(w, 0, peers, nil)
}

func httpSendResponse(w http.ResponseWriter, code int, resp interface{}, err error) {
	if code == 0 {
		// code not specified, determine based on error
		if err == nil {
			code = http.StatusOK
		} else {
			if strings.Contains(err.Error(), "unique constraint") {
				code = http.StatusConflict
			} else if strings.Contains(err.Error(), "Unauthorized") {
				code = http.StatusUnauthorized
			} else {
				code = http.StatusBadRequest
			}
		}
	}

	if isNil(resp) {
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(code)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(resp)
	}
}

func webShowPeers(w http.ResponseWriter, r *http.Request) {
	wgPeers := WgGetPeers()

	fmt.Fprintf(w, headerHtml, listenAddress, listenAddress)
	fmt.Fprintf(w, "<h1><b>Peers</b></h1>")
	fmt.Fprintf(w, tableStyle)

	fmt.Fprintf(w, "<table>")
	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, "<th>Name</th>")
	fmt.Fprintf(w, "<th>Endpoint</th>")
	fmt.Fprintf(w, "<th>Virtual IP</th>")
	fmt.Fprintf(w, "<th>Uptime</th>")
	//fmt.Fprintf(w, "<th>Local IP</th>")
	fmt.Fprintf(w, "<th>Public Key</th>")
	fmt.Fprintf(w, "<th>Applications Allowed</th>")
	fmt.Fprintf(w, "<th>RxBytes</th>")
	fmt.Fprintf(w, "<th>TxBytes</th>")
	fmt.Fprintf(w, "<th>Last Handshake</th>")
	//fmt.Fprintf(w, "<th>Action</th>")
	fmt.Fprintf(w, "</tr>")

	PeersMutex.Lock()
	for _, peer := range Peers {
		if peer.Name == "" {
			continue
		}
		var secs, mins, hrs int
		secs = int(time.Since(peer.StartTime).Seconds())
		if secs >= 60 {
			mins = secs/60
			secs = secs % 60
		}
		if mins >= 60 {
			hrs = mins/60
			mins = mins % 60
		}
		fmt.Fprintf(w, "<tr>")
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", peer.Name))
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", peer.Endpoint))
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", peer.VirtualIP))
		fmt.Fprintf(w, fmt.Sprintf("<td>%d:%02d:%02d</td>", hrs, mins, secs))

		var ip2 string
		/*
		// Local IPs
		for _, ip := range peer.LocalIPs {
			if ip2 != "" {
				ip2 = ip2 + ", "
			}
			ip2 = ip2 + ip
		}
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", ip2))
		*/

		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", peer.WgKey))

		// Allowed IPs
		ip2 = ""
		for _, ip := range peer.AllowedIPs {
			if ip2 != "" {
				ip2 = ip2 + ", "
			}
			ip2 = ip2 + ip
		}
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", ip2))
		wgPeer := wgPeers[peer.WgKey]
		if wgPeer != nil {
			fmt.Fprintf(w, fmt.Sprintf("<td>%d</td>", wgPeer.RxBytes))
			fmt.Fprintf(w, fmt.Sprintf("<td>%d</td>", wgPeer.TxBytes))
			fmt.Fprintf(w, fmt.Sprintf("<td>%d secs ago</td>", int(time.Since(wgPeer.LastHandshakeTime).Seconds())))
		}
		//=============
		//fmt.Fprintf(w, "<td><form action=\"/delete?" + peer.Name + "\"> <input type=\"submit\" value=\"Submit\"></form></td>")
		//=============
		fmt.Fprintf(w, "</tr>")
	}
	PeersMutex.Unlock()

	fmt.Fprintf(w, "<table>")
	fmt.Fprintf(w, footerHtml)
}

func apiDisconnect(w http.ResponseWriter, r *http.Request) {
	Disconnect()
	httpSendResponse(w, 0, nil, nil)
}

func apiLogout(w http.ResponseWriter, r *http.Request) {
	ClientLogout()
	httpSendResponse(w, 0, nil, nil)
}

func apiShowStatus(w http.ResponseWriter, r *http.Request) {
	status := GetStatus()
	httpSendResponse(w, 0, status, nil)
}

func webShowStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, headerHtml, listenAddress, listenAddress)
	fmt.Fprintf(w, "<h1>Status</h1>")


	fmt.Fprintf(w, tableStyle)
	fmt.Fprintf(w, "<table>")

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, fmt.Sprintf("<td>Name</td>"))
	fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", Config.Client.Name))
	fmt.Fprintf(w, "</tr>")

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, fmt.Sprintf("<td>Domain</td>"))
	fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", Config.Client.Domain))
	fmt.Fprintf(w, "</tr>")

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, fmt.Sprintf("<td>Role</td>"))
	fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", Config.Client.Role))
	fmt.Fprintf(w, "</tr>")

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, fmt.Sprintf("<td>Status</td>"))
	fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", Client.Status))

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, fmt.Sprintf("<td>Virtual IP</td>"))
	fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", Client.VirtualIP))
	fmt.Fprintf(w, "</tr>")

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, fmt.Sprintf("<td>Controller Address</td>"))
	fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", Config.Controller.Address))
	fmt.Fprintf(w, "</tr>")

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, fmt.Sprintf("<td>Controller Port</td>"))
	fmt.Fprintf(w, fmt.Sprintf("<td>%d</td>", Config.Controller.Port))
	fmt.Fprintf(w, "</tr>")

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, fmt.Sprintf("<td>Virtual Device</td>"))
	fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", Config.Wireguard.Device))
	fmt.Fprintf(w, "</tr>")

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, fmt.Sprintf("<td>My Public Address</td>"))
	fmt.Fprintf(w, fmt.Sprintf("<td>%s:%d</td>", myPublicIP, WgListenPort))
	fmt.Fprintf(w, "</tr>")

	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, fmt.Sprintf("<td>My Public Key</td>"))
	fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", WgGetMyPublicKey()))
	fmt.Fprintf(w, "</tr>")

	fmt.Fprintf(w, "</table>")

	fmt.Fprintf(w, fmt.Sprintf("<p></p><p></p><p><b>Available Apps</b></p>"))
	fmt.Fprintf(w, tableStyle)
	fmt.Fprintf(w, "<table>")
	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, "<th>Name</th>")
	fmt.Fprintf(w, "<th>Service</th>")
	fmt.Fprintf(w, "<th>Applications Allowed</th>")
	fmt.Fprintf(w, "</tr>")
	for _, app := range Client.Apps {
		fmt.Fprintf(w, "<tr>")
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", app.Name))
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", app.Service))
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", app.AllowedIPs))
		fmt.Fprintf(w, "</tr>")
	}
	fmt.Fprintf(w, "</table>")

	fmt.Fprintf(w, footerHtml)
}

func webShowWireguard(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, headerHtml, listenAddress, listenAddress)
	fmt.Fprintf(w, "<h1>Wireguard</h1>")


	fmt.Fprintf(w, fmt.Sprintf("<p></p><p></p><p><b>Devices</b></p>"))
	fmt.Fprintf(w, tableStyle)
	fmt.Fprintf(w, "<table>")
	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, "<th>Name</th>")
	fmt.Fprintf(w, "<th>Listen Port</th>")
	fmt.Fprintf(w, "<th>Public Key</th>")
	fmt.Fprintf(w, "</tr>")

	devices := WgGetDevices()
	for _, d := range devices {
		fmt.Fprintf(w, "<tr>")
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", d.Name))
		fmt.Fprintf(w, fmt.Sprintf("<td>%d</td>", d.Port))
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", d.PublicKey))
		fmt.Fprintf(w, "</tr>")
	}
	fmt.Fprintf(w, "</table>")

	fmt.Fprintf(w, fmt.Sprintf("<p></p><p></p><p><b>Peers</b></p>"))
	fmt.Fprintf(w, tableStyle)
	fmt.Fprintf(w, "<table>")
	fmt.Fprintf(w, "<tr>")
	fmt.Fprintf(w, "<th>Device</th>")
	fmt.Fprintf(w, "<th>Endpoint</th>")
	fmt.Fprintf(w, "<th>Public Key</th>")
	fmt.Fprintf(w, "<th>Allowed IPs</th>")
	fmt.Fprintf(w, "<th>RxBytes</th>")
	fmt.Fprintf(w, "<th>TxBytes</th>")
	fmt.Fprintf(w, "<th>Last Handshake</th>")
	fmt.Fprintf(w, "</tr>")

	peers := WgGetPeers()
	for _, p := range peers {
		fmt.Fprintf(w, "<tr>")
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", p.DeviceName))
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", p.Endpoint))
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", p.PublicKey))
		fmt.Fprintf(w, fmt.Sprintf("<td>%s</td>", p.AllowedIPs))
		fmt.Fprintf(w, fmt.Sprintf("<td>%d</td>", p.RxBytes))
		fmt.Fprintf(w, fmt.Sprintf("<td>%d</td>", p.TxBytes))
		fmt.Fprintf(w, fmt.Sprintf("<td>%d secs ago</td>", int(time.Since(p.LastHandshakeTime).Seconds())))
		fmt.Fprintf(w, "</tr>")
	}
	fmt.Fprintf(w, "</table>")

	fmt.Fprintf(w, footerHtml)
}

func webServer() {

	// HTML Response
	http.HandleFunc("/", webShowStatus)
	http.HandleFunc("/status", webShowStatus)
	http.HandleFunc("/peers", webShowPeers)
	http.HandleFunc("/wireguard", webShowWireguard)

	// JSON Response
	http.HandleFunc("/api/status", apiShowStatus)
	http.HandleFunc("/api/peers", apiShowPeers)
	http.HandleFunc("/api/disconnect", apiDisconnect)
	http.HandleFunc("/api/logout", apiLogout)

	listenAddress = fmt.Sprintf("localhost:%d", Config.Webserver.Port)
	log.Printf("Starting web server on %s\n", listenAddress)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

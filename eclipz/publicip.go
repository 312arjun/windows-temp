package eclipz

import (
	"log"
	"net/http"
	"io/ioutil"
)

var myPublicIP string

func myPublicAddress() string {
	url := "http://api.ipify.org?format=text"
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Unable to get MyPublicIP from %s\n", url)
		return ""
	}
	ip, err := ioutil.ReadAll(resp.Body)
	if err == nil {
		resp.Body.Close()
		log.Printf("My Public IP: %s\n", ip)
		myPublicIP = string(ip)
		return myPublicIP
	}
	resp.Body.Close()

	// to get public IP through another site
	log.Printf("Unable to get MyPublicIP from %s\n", url)
	return ""
}

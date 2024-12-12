package eclipz

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"bufio"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

var ClientReadyNotification = make(chan bool)
var ErrorChannel = make(chan string)
var RetryControllerConnChannel = make(chan bool)
var RequestCredentials = make(chan bool)

func ClientLogin() {
	req := &RpcLoginReq{}
	req.Cmd = CmdLoginReq
	req.MsgId = getNextMsgId()
	req.Name = Config.Client.Name
	req.Domain = Config.Client.Domain
	req.Password = Config.Client.Password
	req.Role = Config.Client.Role
	req.DeviceId = GetMyDeviceId()
	ClientStatusNotification <- GetStatus()

	sendQueue <- req
}

func processLoginResp(msg []byte) error {
	var resp RpcLoginResp
	err := json.Unmarshal(msg, &resp)
	if err != nil {
		msg := "Data is not JSON\n"
		log.Printf(msg)
		ErrorChannel <- msg
		return err
	}
	if resp.Status != StatusSuccess {
		msg := fmt.Sprintf("An error occurred during login: %s\n", resp.StatusText)
		ErrorChannel <- msg
		return fmt.Errorf(msg)
	}
	resetRetryInterval()
	log.Printf("LoginResponse; Status=%d, Token=%s\n", resp.Status, resp.Token)
	Client.AuthToken = resp.Token
	Client.WgControllerKey = resp.WgControllerKey
	Client.ChangeStatus(STATUS_LOGIN_SUCCESS)

	if Client.WgControllerKey != "" {
		peer := configureBackendWireguard(Client.WgControllerKey)

		// Wait for ClientReady notification from backend before sending ClientInfo
		select {
		case <-ClientReadyNotification:
		case <-time.After(15 * time.Second):
		}

		// Wireguard configuration for backend peer can be removed now
		WgRemovePeer(peer)
	}
	ClientStatusNotification <- GetStatus()
	sendClientInfo()
	return nil
}

func ClientLogout() {
	req := &RpcLogoutReq{}
	req.Cmd = CmdLogoutReq
	req.MsgId = getNextMsgId()
	req.Token = Client.AuthToken

	sendQueue <- req
}

func processLogoutResp(msg []byte) error {
	var resp RpcCommon
	err := json.Unmarshal(msg, &resp)
	if err != nil {
		log.Printf("Data is not JSON\n")
		return err
	}
	if resp.Status != StatusSuccess {
		log.Printf("Logout Failed: %d %s\n", resp.Status, resp.StatusText)
	}
	return nil
}

func readCredentials() error {
	if Config.Client.Name != "" && Config.Client.Domain != "" &&
		Config.Client.Password != "" {
		return nil
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	fmt.Print("Enter Domain: ")
	domain, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return err
	}

	username = strings.TrimSpace(username)
	domain = strings.TrimSpace(domain)
	password := strings.TrimSpace(string(bytePassword))

	// Save creadentials in config file
	Config.Client.Name = username
	Config.Client.Domain = domain
	Config.Client.Password = password

	return nil
}

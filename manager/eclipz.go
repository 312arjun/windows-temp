package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/sys/windows"

	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"golang.zx2c4.com/wireguard/windows/eclipz"
	"golang.zx2c4.com/wireguard/windows/l18n"
	"golang.zx2c4.com/wireguard/windows/tunnel"
)

var eclipzHomeDir = "C:\\Program Files\\Eclipz"

func fatalError(msg string) {
	windows.MessageBox(0, windows.StringToUTF16Ptr(msg),
		windows.StringToUTF16Ptr(l18n.Sprintf("Error")),
		windows.MB_ICONERROR)
	err := UninstallManager()
	if err == nil {
		os.Exit(1)
	}
}

func CheckConfigFile() bool {
	// Prompt a file selection
	var configFile string = eclipzHomeDir + "\\config.json"
	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

func updateEclipzConfig(privKey string, pubKey string) {
	var configFile string = eclipzHomeDir + "\\config.json"
	var config eclipz.Configuration

	content, err := os.ReadFile(configFile)
	if err != nil {
		fatalError(fmt.Sprintf("Unable to read %s\n", configFile))
	}

	err = json.Unmarshal(content, &config)
	if err != nil {
		fatalError(fmt.Sprintf("File %s - Not a valid JSON file\n", configFile))
	}

	// Update config file
	config.Wireguard.PrivateKey = privKey
	config.Wireguard.PublicKey = pubKey

	// Write config file
	var data []byte
	data, _ = json.MarshalIndent(&config, "", "    ")
	_ = ioutil.WriteFile(configFile, data, 0600)

	tunnel.WGlog("Added Private & Public Key %s to file %s", pubKey, configFile)
}

func createWgConfig(path string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		tunnel.WGlog("Error while creating: %v", err)
		return err
	}
	privKey, _ := wgtypes.GeneratePrivateKey()
	privKeyStr := privKey.String()

	text := fmt.Sprintf("[Interface]\nPrivateKey=%s\n", privKeyStr)
	_, _ = file.WriteString(text)
	file.Close()
	tunnel.WGlog("Created file %s", path)

	pubKeyStr := privKey.PublicKey().String()
	updateEclipzConfig(privKeyStr, pubKeyStr)
	return nil
}

// Install the eclipz0.conf file we need if it does not already exist.
func InstallConfFile(name string) {
	path := eclipzHomeDir + "\\Data\\Configurations\\" + name + ".conf"
	ePath := path + ".dpapi"
	tunnel.WGlog("InstallConfFile: path: %s, ePath: %s \n", path, ePath)

	_, err := os.Stat(ePath) // check for presence of encrypted conf file
	if err == nil {
		// Use already existing file
		tunnel.WGlog("InstallConfFile: using existing file; ePath: %s \n", ePath)
		return
	}

	// Create path for unencrypted conf file. WGW will encrypt it
	_, err = os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		tunnel.WGlog("InstallConfFile: calling createWgConfig(%s) \n", path)
		err = createWgConfig(path)
	}
	if err != nil {
		// Conf file does not exist and unable to create
		tunnel.WGlog("InstallConfFile() unable to create %s: %v", path, err)
		errstr := fmt.Sprintf("InstallConfFile() unable to create %s: %v", path, err)
		fatalError(errstr)
	}
}

// Wrapper for install tunnel which make the full Eclipz path from the tunnel iface name.
func InstallTunnelByName(name string) {
	InstallConfFile(name)
	path := eclipzHomeDir + "\\Data\\Configurations\\" + name + ".conf.dpapi"
	err := InstallTunnel(path)
	if isTunnelRunning(err) {
		tunnel.WGlog("InstallTunnelByName(); path: %s error: %v", path, err)
		return
	}
	if err != nil {
		errstr := fmt.Sprintf("InstallTunnelByName() unable to open tunnel %s; %v", path, err)
		//fatalError(errstr)
		tunnel.WGlog(errstr)
		err := UninstallManager()
		if err == nil {
			os.Exit(1)
		}
	}
	tunnel.WGlog("InstallTunnelByName(); path: %s error: %v", path, err)
}

func DeleteConfigFile() error {
	// Specify the file path you want to delete
	var configFile = eclipzHomeDir + "\\config.json"

	// Attempt to delete the file
	err := os.Remove(configFile)
	if err != nil {
		return err
	}
	return nil
}

func isTunnelRunning(err error) bool {
	return err.Error() == "Tunnel already installed and running"
}

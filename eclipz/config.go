package eclipz

import (
	"encoding/json"
	"log"
	"os"
)

type ControllerConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	CACert  string `json:"cacert"`
}

type ClientConfig struct {
	Name           string `json:"name"`
	Domain         string `json:"domain"`
	Device         string `json:"device,omitempty"`
	Password       string `json:"password"`
	PrevPassword   string `json:"prev_password"`
	Role           string `json:"role"`
	StatsInterval  int    `json:"stats_interval"`
	LogFile        string `json:"log_file,omitempty"`
	LogLevel       string `json:"log_level,omitempty"`
	ConnName       string `json:"conn_name,omitempty"`
	Enable         bool   `json:"enable,omitempty"`
	DhcpPoolEnable bool   `json:"dhcp_pool_enable,omitempty"`
}

type WebserverConfig struct {
	Enable  bool   `json:"enable"`
	Address string `json:"address"`
	Port    int    `json:"port"`
}

type WgConfig struct {
	Device     string `json:"device"`
	Address    string `json:"address"`
	Port       int    `json:"port"`
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	Keepalive  int    `json:"keepalive,omitempty"`
	FWmark     int    `json:"fwmark,omitempty"`
	LogLevel   string `json:"log_level,omitempty"`
}

type Configuration struct {
	Controller ControllerConfig `json:"controller"`
	Client     ClientConfig     `json:"client"`
	Wireguard  WgConfig         `json:"wireguard"`
	Webserver  WebserverConfig  `json:"webserver"`
}

var Config Configuration

var ConfigFile string
var AltConfigFile string = "config.json"

func SetConfigFile(filename string) {
	ConfigFile = filename
	log.Printf("Using Config %s\n", ConfigFile)
}

// Update configuration changes
func UpdateConfigJSON(data string) error {
	bytes := []byte(data)
	config := Configuration{}
	err := json.Unmarshal(bytes, &config)
	if err != nil {
		return err
	}
	UpdateConfig(&config)
	return nil
}

// Update configuration changes
func UpdateConfig(config *Configuration) {
	log.Printf("Update Config\n")
}

func loadConfig() {
	if ConfigFile == "" {
		ConfigFile = getConfigFileName()
	}

	content, err := os.ReadFile(ConfigFile)
	if err != nil {
		content, err = os.ReadFile(AltConfigFile)
		if err != nil {
			log.Fatalf("Unable to read either %s or %s: %v\n", ConfigFile, AltConfigFile, err)
			EClog("Reading error: %+v", err)
		}
	}

	err = json.Unmarshal(content, &Config)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}
	log.Printf("Config: %+v\n", Config)
	EClog("Config: %+v\n", Config)
}

var ConfigChanged bool

func configChanged() {
	ConfigChanged = true
}

func saveConfigFile() {
	if !ConfigChanged {
		return
	}
	var data []byte
	data, _ = json.MarshalIndent(&Config, "", "    ")

	_ = os.WriteFile(ConfigFile, data, 0600)
	ConfigChanged = false
}

func CreateAdapterDefaultConfigFile() Configuration {
	return Configuration{
		Controller: ControllerConfig{
			CACert: "cachain.pem",
		},
		Client: ClientConfig{
			Role:          DB_ROLE_CLIENT,
			StatsInterval: 60,
		},
		Wireguard: WgConfig{
			Device:     "eclipz0",
			Address:    "",
			Port:       51820,
			PrivateKey: "",
			PublicKey:  "",
		},
		Webserver: WebserverConfig{
			Enable:  true,
			Address: "localhost",
			Port:    8088,
		},
	}
}

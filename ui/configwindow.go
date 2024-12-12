package ui

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/lxn/walk"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	// . "github.com/lxn/walk/declarative"
)

type ConfigWindow struct {
	walk.FormBase
	result bool
}

type ControllerConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
	CACert  string `json:"cacert"`
}

type ClientConfig struct {
	Name           string `json:"name"`
	Domain         string `json:"domain"`
	Device         string `json:"device"`
	Password       string `json:"password"`
	PrevPassword   string `json:"prev_password"`
	Role           string `json:"role"`
	StatsInterval  int    `json:"stats_interval"`
	LogFile        string `json:"log_file"`
	LogLevel       string `json:"log_level"`
	ConnName       string `json:"conn_name"`
	Enable         bool   `json:"enable"`
	DhcpPoolEnable bool   `json:"dhcp_pool_enable"`
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
	Keepalive  int    `json:"keepalive"`
	FWmark     int    `json:"fwmark"`
	LogLevel   string `json:"log_level"`
}

type Configuration struct {
	Controller ControllerConfig `json:"controller"`
	Client     ClientConfig     `json:"client"`
	Wireguard  WgConfig         `json:"wireguard"`
	Webserver  WebserverConfig  `json:"webserver"`
}

var Config Configuration

const (
	winClassName = "Custom Dialog"
)

var selectConfigFile sync.Once

func NewConfigWindow() (*ConfigWindow, error) {
	selectConfigFile.Do(func() {
		walk.AppendToWalkInit(func() {
			walk.MustRegisterWindowClass(winClassName)
		})
	})

	var err error

	cw := new(ConfigWindow)
	cw.SetName("Config File")

	err = walk.InitWindow(cw, nil, winClassName, win.WS_OVERLAPPEDWINDOW, win.WS_EX_CONTROLPARENT)
	if err != nil {
		return nil, err
	}

	cw.SetPersistent(true)

	cw.SetMinMaxSize(walk.Size{Width: 500, Height: 250}, walk.Size{Width: 0, Height: 0})
	vlayout := walk.NewVBoxLayout()
	cw.SetLayout(vlayout)

	var (
		com0 *walk.Composite
		com1 *walk.Composite
		com2 *walk.Composite
		iv   *walk.ImageView
		t1   *walk.TextLabel
		t2   *walk.TextLabel
		b1   *walk.PushButton
		b2   *walk.PushButton
		b3   *walk.PushButton
	)

	com0, err = walk.NewComposite(cw)
	if err != nil {
		return nil, err
	}
	com0.SetLayout(walk.NewVBoxLayout())
	com0.SetMinMaxSize(walk.Size{Width: 500, Height: 65}, walk.Size{Width: 500, Height: 65})

	iv, _ = walk.NewImageView(com0)
	img, err := walk.NewImageFromFileForDPI("C:\\Program Files\\Eclipz\\img\\Banner.bmp", 96)
	if err == nil {
		iv.SetImage(img)
	}

	com1, err = walk.NewComposite(cw)
	if err != nil {
		return nil, err
	}
	com1.SetLayout(walk.NewVBoxLayout())
	com1.SetMinMaxSize(walk.Size{Width: 500, Height: 300}, walk.Size{Width: 500, Height: 300})

	t1, _ = walk.NewTextLabel(com1)
	t1.SetText("Configuration File Missing")
	font, err := walk.NewFont("Segoe UI", 14, walk.FontBold)
	if err != nil {
		return nil, err
	}
	t1.SetFont(font)
	t1.SetTextAlignment(walk.AlignHCenterVNear)
	t1.SetMinMaxSize(walk.Size{Width: 100, Height: 100}, walk.Size{Width: 100, Height: 100})
	t1.SetAlignment(walk.AlignHCenterVNear)

	t2, _ = walk.NewTextLabel(com1)
	t2.SetText("Select the JSON configuration file that your administrator emailed to you. Contact your administrator if you do not have this file.")
	font, err = walk.NewFont("Segoe UI", 12, 0)
	if err != nil {
		return nil, err
	}
	t2.SetFont(font)

	com2, err = walk.NewComposite(cw)
	if err != nil {
		return nil, err
	}
	com2.SetLayout(walk.NewHBoxLayout())
	com2.SetAlignment(walk.AlignHCenterVFar)
	com2.SetMinMaxSize(walk.Size{Width: 500, Height: 80}, walk.Size{Width: 0, Height: 0})

	b1, _ = walk.NewPushButton(com2)
	b1.SetText("Exit")
	font, err = walk.NewFont("Segoe UI", 10, 0)
	if err != nil {
		return nil, err
	}
	b1.SetFont(font)
	b1.SetMinMaxSize(walk.Size{Width: 105, Height: 40}, walk.Size{Width: 105, Height: 40})
	b1.Clicked().Attach(cw.b1Handler)

	b2, _ = walk.NewPushButton(com2)
	b2.SetText("Select JSON File")
	b2.SetFont(font)
	b2.SetMinMaxSize(walk.Size{Width: 105, Height: 40}, walk.Size{Width: 105, Height: 40})
	b2.Clicked().Attach(cw.b2Handler)

	b3, _ = walk.NewPushButton(com2)
	b3.SetText("Set Manually")
	b3.SetFont(font)
	b3.SetMinMaxSize(walk.Size{Width: 105, Height: 40}, walk.Size{Width: 105, Height: 40})
	b3.Clicked().Attach(cw.b3Handler)

	// // Remove caption bar and center the windows
	// defaultStyle := win.GetWindowLong(cw.Handle(), win.GWL_STYLE) // Gets current style
	// newStyle := defaultStyle &^ win.WS_CAPTION                    // Remove WS_THICKFRAME
	// win.SetWindowLong(cw.Handle(), win.GWL_STYLE, newStyle)

	win.SetWindowLong(cw.Handle(), win.GWL_STYLE, win.GetWindowLong(cw.Handle(), win.GWL_STYLE) & ^win.WS_MAXIMIZEBOX & ^win.WS_SIZEBOX)

	xScreen := win.GetSystemMetrics(win.SM_CXSCREEN)
	yScreen := win.GetSystemMetrics(win.SM_CYSCREEN)
	win.SetWindowPos(
		cw.Handle(),
		0,
		(xScreen-500)/2,
		(yScreen-300)/3,
		500,
		300,
		win.SWP_FRAMECHANGED,
	)

	return cw, nil
}

func (cw *ConfigWindow) b1Handler() {
	log.Fatal(fmt.Errorf("missing config.json"))
}

func (cw *ConfigWindow) b2Handler() {
	configFile, err := cw.openFile()
	if err != nil || configFile == "" {
		return
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		msg := "Can't read the file"
		windows.MessageBox(0, windows.StringToUTF16Ptr(msg), windows.StringToUTF16Ptr("Error"), windows.MB_ICONERROR)
		return
	}

	err = saveConfigFile(content)

	if err != nil {
		err = errors.New("Error while copying the configuration file")
		showErrorPopup(err)
		return
	}

	cw.Close()
}

func showErrorPopup(err error) {
	fmt.Printf("Err: %+v", err)
	windows.MessageBox(0, windows.StringToUTF16Ptr(err.Error()), windows.StringToUTF16Ptr("Error"), windows.MB_ICONERROR)
}

func saveConfigFile(content []byte) error {
	dstFile := "C:\\Program Files\\Eclipz\\config.json"
	err := os.WriteFile(dstFile, content, 0644)
	return err
}

func (cw *ConfigWindow) b3Handler() {
	NewManualConfig(cw)
	cw.Close()
}

func (cw *ConfigWindow) Dispose() {

	cw.FormBase.Dispose()
}

func (cw *ConfigWindow) openFile() (file string, err error) {
	dlgFile := new(walk.FileDialog)
	dlgFile.Filter = "Config file(*.json)|*.json"
	dlgFile.Title = "Config File"

	if ok, err := dlgFile.ShowOpen(cw); err != nil {
		return dlgFile.FilePath, err
	} else if !ok {
		return dlgFile.FilePath, err
	}

	file = dlgFile.FilePath

	return file, nil
}

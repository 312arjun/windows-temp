package ui

import (
	"encoding/json"
	"log"
	"strconv"
	"unicode"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.zx2c4.com/wireguard/windows/eclipz"
	"golang.zx2c4.com/wireguard/windows/manager"
	"golang.zx2c4.com/wireguard/windows/tunnel"
)

type ManualConfig struct {
	dlg                *walk.Dialog
	usernameLineEdit   *walk.LineEdit
	passwordLineEdit   *walk.LineEdit
	domainLineEdit     *walk.LineEdit
	controllerLineEdit *walk.LineEdit
	portLineEdit       *walk.LineEdit

	defaultPortCheckBox *walk.CheckBox
	showPasswordButton  *walk.PushButton
	loginButton         *walk.PushButton
}

const DEFAULT_PORT = "8443"

func NewManualConfig(parent walk.Container) {
	manualConfig := new(ManualConfig)
	var username, password, domain, controller, port string
	configExists := manager.CheckConfigFile()
	if configExists {
		username = eclipz.Config.Client.Name
		controller = eclipz.Config.Controller.Address
		password = eclipz.Config.Client.Password
		domain = eclipz.Config.Client.Domain
		port = strconv.Itoa(eclipz.Config.Controller.Port)
	}
	_, err := Dialog{
		Persistent: true,
		Title:      "Update your credentials to connect",
		MinSize:    Size{Width: 400},
		Layout:     VBox{},
		AssignTo:   &manualConfig.dlg,
		Children: []Widget{
			Composite{
				Layout: HBox{SpacingZero: true},
				Children: []Widget{
					Composite{
						Layout: VBox{SpacingZero: true, MarginsZero: true},
						Children: []Widget{
							Label{Text: "Username:"},
							LineEdit{
								AssignTo: &manualConfig.usernameLineEdit,
								Text:     username,
							},
						},
					},
					Composite{
						Layout: VBox{SpacingZero: true, MarginsZero: true},
						Children: []Widget{
							Label{Text: "Domain:"},
							LineEdit{
								AssignTo: &manualConfig.domainLineEdit,
								Text:     domain,
							},
						},
					},
				},
			},
			Composite{
				Layout: HBox{SpacingZero: true},
				Children: []Widget{
					Composite{
						Layout: VBox{SpacingZero: true, MarginsZero: true},
						Children: []Widget{
							Label{Text: "Password:"},
							Composite{
								Layout: HBox{SpacingZero: true, MarginsZero: true},
								Children: []Widget{
									LineEdit{
										AssignTo:     &manualConfig.passwordLineEdit,
										PasswordMode: true,
										Text:         password,
									},
									PushButton{
										AssignTo: &manualConfig.showPasswordButton,
										MinSize:  Size{Width: 90},
										Text:     "Show password",
										OnClicked: func() {
											show := !manualConfig.passwordLineEdit.PasswordMode()
											manualConfig.passwordLineEdit.SetPasswordMode(show)
											manualConfig.passwordLineEdit.SetFocus()
											var btnText string
											if show {
												btnText = "Show password"
											} else {
												btnText = "Hide password"
											}
											manualConfig.showPasswordButton.SetText(btnText)
										},
									},
								},
							},
						},
					},
				},
			},
			Composite{
				Layout: HBox{SpacingZero: true},
				Children: []Widget{
					Composite{
						Layout: VBox{SpacingZero: true, MarginsZero: true},
						Children: []Widget{
							Label{Text: "Controller:"},
							Composite{
								Layout: HBox{SpacingZero: true, MarginsZero: true},
								Children: []Widget{
									LineEdit{
										AssignTo: &manualConfig.controllerLineEdit,
										Text:     controller,
									},
									CheckBox{
										AssignTo: &manualConfig.defaultPortCheckBox,
										Text:     "Default port",
										OnCheckedChanged: func() {
											if manualConfig.defaultPortCheckBox.Checked() {
												manualConfig.portLineEdit.SetText(DEFAULT_PORT)
												manualConfig.portLineEdit.SetEnabled(false)
											} else {
												manualConfig.portLineEdit.SetEnabled(true)
											}
										},
									},
								},
							},
						},
					},
				},
			},
			Composite{
				Layout: VBox{SpacingZero: true},
				Children: []Widget{
					Label{Text: "Port:"},
					LineEdit{
						AssignTo: &manualConfig.portLineEdit,
						Text:     port,
					},
				},
			},
			PushButton{
				AssignTo: &manualConfig.loginButton,
				Text:     "Login",
				OnClicked: func() {
					defer func() {
						if r := recover(); r != nil {
							tunnel.WGlog("Recovered in %v", r)
						}
					}()

					if err := validateNumberLineEdit(manualConfig.portLineEdit); err != nil {
						newErrorMessage(manualConfig.dlg, "Invalid port number.")
						return
					}
					username := manualConfig.usernameLineEdit.Text()
					password := manualConfig.passwordLineEdit.Text()
					domain := manualConfig.domainLineEdit.Text()
					port := manualConfig.portLineEdit.Text()
					controller := manualConfig.controllerLineEdit.Text()

					config := eclipz.CreateAdapterDefaultConfigFile()
					intport, _ := strconv.Atoi(port)
					config.Controller.Port = intport
					config.Controller.Address = controller
					config.Client.Domain = domain
					config.Client.Name = username
					config.Client.Password = password
					if controller == "" || domain == "" || username == "" || password == "" {
						newErrorMessage(manualConfig.dlg, "Some fields are empty. Please check the form.")
						return
					}

					data, err := json.Marshal(config)
					if err != nil {
						newErrorMessage(manualConfig.dlg, "Invalid config. Cannot create a JSON file.")
						return
					}

					err = saveConfigFile(data)
					if err != nil {
						newErrorMessage(manualConfig.dlg, "Error saving the config file.")
						return
					}
					if configExists {
						eclipz.RetryControllerConnChannel <- true
					}
					manualConfig.dlg.Close(walk.DlgCmdOK)

				},
			},
		},
	}.Run(parent.Form())
	if err != nil {
		log.Fatal(err)
	}
}

func NumericValidator(text string) (bool, error) {
	for _, r := range text {
		if !unicode.IsNumber(r) {
			return false, nil
		}
	}
	return true, nil
}

func validateNumberLineEdit(lineEdit *walk.LineEdit) error {
	text := lineEdit.Text()
	_, err := strconv.ParseFloat(text, 64)
	return err
}

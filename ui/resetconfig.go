package ui

import (
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.zx2c4.com/wireguard/windows/eclipz"
	"golang.zx2c4.com/wireguard/windows/manager"
)

type ResetConfigWindow struct {
	dlg *walk.Dialog
}

func newResetConfigWindow(parent walk.Form) {
	resetConfig := new(ResetConfigWindow)
	_, err := Dialog{
		Persistent: true,
		Visible:    true,
		Title:      "Reset config",
		MinSize:    Size{Width: 300},
		Layout:     VBox{},
		Children: []Widget{
			Label{
				Text: "Are you sure you want to reset config file?",
			},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					PushButton{
						Text:    "Close",
						MinSize: Size{Width: 100},
						OnClicked: func() {
							resetConfig.dlg.Cancel()
						},
					},
					PushButton{
						Text:    "Accept",
						MinSize: Size{Width: 100},
						OnClicked: func() {
							defer resetConfig.dlg.Accept()
							// Specify the file path you want to delete
							err := manager.DeleteConfigFile()
							if err != nil {
								eclipz.ErrorChannel <- err.Error()
							}
							parent.Dispose()
						},
					},
				},
			},
		},
		AssignTo: &resetConfig.dlg,
	}.Run(parent)
	if err != nil {
		log.Fatal(err)
	}
}

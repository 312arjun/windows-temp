package ui

import (
	"log"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type ErrorMessage struct {
	dlg *walk.Dialog
}

func newErrorMessage(parent walk.Form, message string) {
	errorWindow := new(ErrorMessage)
	_, err := Dialog{
		Persistent: true,
		Visible:    true,
		Title:      "Error",
		MinSize:    Size{Width: 300},
		Layout:     VBox{},
		Children: []Widget{
			Label{
				Text: message,
			},
			PushButton{
				Text: "Close",
				OnClicked: func() {
					errorWindow.dlg.Cancel()
				},
			},
		},
		AssignTo: &errorWindow.dlg,
	}.Run(parent)
	if err != nil {
		log.Fatal(err)
	}
}

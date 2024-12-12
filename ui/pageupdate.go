/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.zx2c4.com/wireguard/windows/l18n"
	"golang.zx2c4.com/wireguard/windows/manager"
	"golang.zx2c4.com/wireguard/windows/updater"
)

type UpdatePage2 struct {
	*walk.Composite

	t1  *walk.TextLabel
	bar *walk.ProgressBar
	btn *walk.PushButton
}

func NewUpdatePage2(parent walk.Container) (Page, error) {
	p := new(UpdatePage2)

	if err := (Composite{
		AssignTo: &p.Composite,
		Name:     "UpdatePage",
		Layout:   VBox{},
		Children: []Widget{
			TextLabel{
				Text:    "An update to Eclipz is available. It is highly advisable to update without delay.",
				MinSize: Size{Width: 630, Height: 0},
			},
			TextLabel{
				Text:     "Status: Waiting for user",
				MinSize:  Size{Width: 1, Height: 0},
				AssignTo: &p.t1,
			},
			ProgressBar{
				Visible:  false,
				AssignTo: &p.bar,
			},
			PushButton{
				Text:     l18n.Sprintf("Update Now"),
				MaxSize:  Size{Width: 150, Height: 50},
				AssignTo: &p.btn,
			},
			VSpacer{},
		},
	}).Create(NewBuilder(parent)); err != nil {
		return nil, err
	}

	if err := walk.InitWrapperWindow(p); err != nil {
		return nil, err
	}

	if !IsAdmin {
		p.btn.SetText(l18n.Sprintf("Please ask the system administrator to update."))
		p.btn.SetEnabled(false)
		p.btn.SetText(l18n.Sprintf("Status: Waiting for administrator"))
	}

	switchToUpdatingState := func() {
		if !p.bar.Visible() {
			p.SetSuspended(true)
			p.btn.SetEnabled(false)
			p.btn.SetVisible(false)
			p.bar.SetVisible(true)
			p.bar.SetMarqueeMode(true)
			p.SetSuspended(false)
			p.t1.SetText(l18n.Sprintf("Status: Waiting for updater service"))
		}
	}

	switchToReadyState := func() {
		if p.bar.Visible() {
			p.SetSuspended(true)
			p.bar.SetVisible(false)
			p.bar.SetValue(0)
			p.bar.SetRange(0, 1)
			p.bar.SetMarqueeMode(false)
			p.btn.SetVisible(true)
			p.btn.SetEnabled(true)
			p.SetSuspended(false)
		}
	}

	p.btn.Clicked().Attach(func() {
		switchToUpdatingState()
		err := manager.IPCClientUpdate()
		if err != nil {
			switchToReadyState()
			p.t1.SetText(l18n.Sprintf("Error: %v. Please try again.", err))
		}
	})

	manager.IPCClientRegisterUpdateProgress(func(dp updater.DownloadProgress) {
		p.Synchronize(func() {
			switchToUpdatingState()
			if dp.Error != nil {
				switchToReadyState()
				err := dp.Error
				p.t1.SetText(l18n.Sprintf("Error: %v. Please try again.", err))
				return
			}
			if len(dp.Activity) > 0 {
				stateText := dp.Activity
				p.t1.SetText(l18n.Sprintf("Status: %s", stateText))
			}
			if dp.BytesTotal > 0 {
				p.bar.SetMarqueeMode(false)
				p.bar.SetRange(0, int(dp.BytesTotal))
				p.bar.SetValue(int(dp.BytesDownloaded))
			} else {
				p.bar.SetMarqueeMode(true)
				p.bar.SetValue(0)
				p.bar.SetRange(0, 1)
			}
			if dp.Complete {
				switchToReadyState()
				p.t1.SetText(l18n.Sprintf("Status: Complete!"))
				return
			}
		})
	})

	return p, nil
}

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.zx2c4.com/wireguard/windows/l18n"
	"golang.zx2c4.com/wireguard/windows/version"
)

type AboutPage struct {
	*walk.Composite
	iv *walk.ImageView
}

func newAboutPage(parent walk.Container) (Page, error) {
	p := new(AboutPage)

	if err := (Composite{
		AssignTo: &p.Composite,
		Name:     "AboutPage",
		Layout:   VBox{},
		Children: []Widget{
			ImageView{
				AssignTo: &p.iv,
				MinSize:  Size{Width: 630, Height: 70},
			},
			TextLabel{
				TextColor:     TEXT_WHITE,
				TextAlignment: AlignHCenterVNear,
				Text:          "Eclipz Agent for Windows (64-bits)",
			},
			TextLabel{
				TextColor:     TEXT_WHITE,
				TextAlignment: AlignHCenterVNear,
				Text:          version.Number,
			},
			TextLabel{
				TextColor:     TEXT_WHITE,
				TextAlignment: AlignHCenterVNear,
				Text:          "Copyright 2022 Eclipz, Inc",
			},
		},
	}).Create(NewBuilder(parent)); err != nil {
		return nil, err
	}

	if err := walk.InitWrapperWindow(p); err != nil {
		return nil, err
	}

	if logo, err := loadLogoIconWhite(128); err == nil {
		p.iv.SetImage(logo)
	}
	p.iv.Accessibility().SetName(l18n.Sprintf("Eclipz logo image"))

	p.SetSize(walk.Size{Width: 500, Height: 250})

	return p, nil
}

func (mw *AgentWindow) buttonHandler4() {
	if prevPage := mw.currentPage; prevPage != nil {
		mw.pageCom.SaveState()
		prevPage.SetVisible(false)
		prevPage.(walk.Widget).SetParent(nil)
		prevPage.Dispose()
	}

	page, err := newAboutPage(mw.pageCom)
	if err != nil {
		return
	}

	mw.currentPage = page
}

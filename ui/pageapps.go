/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.zx2c4.com/wireguard/windows/eclipz"
	"golang.zx2c4.com/wireguard/windows/tunnel"
)

type AppsPage struct {
	*walk.Composite
	tbView *walk.TableView
}

type App struct {
	Name       string `json:"name,omitempty"`
	Service    string `json:"service,omitempty"`
	AllowedIPs string `json:"allowed_ips,omitempty"`
}

type AppsModel struct {
	walk.TableModelBase

	items []*App
}

func NewAppsModel() *AppsModel {
	m := new(AppsModel)
	maxrows := 10

	// Get apps from eclipz module
	status := eclipz.GetStatus()

	// Make model items
	m.items = make([]*App, maxrows)

	for i, app := range status.Apps {
		m.items[i] = &App{Name: app.Name, Service: app.Service, AllowedIPs: app.AllowedIPs}
	}

	return m
}

func (m *AppsModel) RowCount() int {
	return len(m.items)
}

func (m *AppsModel) Value(row int, col int) interface{} {
	item := m.items[row]
	if item == nil {
		return ""
	}

	switch col {
	case 0:
		return item.Name

	case 1:
		return item.Service

	case 2:
		return item.AllowedIPs
	}

	panic("unexpected col")
}

func newAppsPage(parent walk.Container) (Page, error) {
	defer func() {
		if r := recover(); r != nil {
			tunnel.WGlog("Error Error Error Error: %v", r)
		}
	}()
	p := new(AppsPage)

	if err := (Composite{
		AssignTo:   &p.Composite,
		Name:       "AppsPage",
		Background: SolidColorBrush{Color: BACKGROUND_COLOR},
		Layout:     VBox{},
		Children: []Widget{
			Composite{
				Layout:  HBox{},
				MaxSize: Size{Width: 0, Height: 295},
				Children: []Widget{
					TableView{
						AssignTo:            &p.tbView,
						Background:          SolidColorBrush{Color: BACKGROUND_TABLE},
						LastColumnStretched: true, // Stretch the last column to fill the available space
						StretchFactor:       1,    // Stretch the TableView itself to fill the available space
						Columns: []TableViewColumn{
							{Title: "Name", Width: 150},
							{Title: "Gateway", Width: 150},
							{Title: "Application Allowed", Width: 200},
						},
					},
				},
			},
			Composite{
				Background: SolidColorBrush{Color: BACKGROUND_COLOR},
				Layout:     HBox{},
				MaxSize:    Size{Width: 500, Height: 100},
			},
		},
	}).Create(NewBuilder(parent)); err != nil {
		return nil, err
	}

	if err := walk.InitWrapperWindow(p); err != nil {
		return nil, err
	}

	p.tbView.SetCellStyler(p)

	// Set model to table view
	model := NewAppsModel()
	if model != nil {
		p.tbView.SetModel(model)
	}

	return p, nil
}

func (s *AppsPage) StyleCell(style *walk.CellStyle) {
	style.BackgroundColor = BACKGROUND_TABLE
	style.TextColor = TEXT_WHITE
}

func (s *AppsPage) StyleHeader(style *walk.CellStyle) {
	style.BackgroundColor = BACKGROUND_TABLE
	style.TextColor = TEXT_WHITE
}

func (mw *AgentWindow) buttonHandler2() {
	if prevPage := mw.currentPage; prevPage != nil {
		mw.pageCom.SaveState()
		prevPage.SetVisible(false)
		prevPage.(walk.Widget).SetParent(nil)
		prevPage.Dispose()
	}

	page, err := newAppsPage(mw.pageCom)
	if err != nil {
		return
	}

	mw.currentPage = page
}

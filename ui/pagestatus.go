/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"sync"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"golang.zx2c4.com/wireguard/windows/eclipz"
	"golang.zx2c4.com/wireguard/windows/tunnel"
)

type StatusPage struct {
	*walk.Composite
	tbView *walk.TableView
}

type Status struct {
	Key   string
	Value string
}

type StatusModel struct {
	walk.TableModelBase
	data []Status
	sync.RWMutex
}

func (m *StatusModel) SetValue(row, col int, value interface{}) error {
	m.Lock()
	defer m.Unlock()
	data := m.data[row]
	tunnel.WGlog("Data to change: %s to %s", data.Value, value)

	switch col {
	case 0:
		data.Key = value.(string)
	case 1:
		data.Value = value.(string)
	default:
		panic("unexpected column")
	}

	m.PublishRowChanged(row)

	return nil
}

func NewStatusModel() *StatusModel {
	m := new(StatusModel)
	// Get status from eclipz module
	status := eclipz.GetStatus()

	// Make model items
	m.data = make([]Status, 11)

	m.data[0] = Status{Key: "Name", Value: status.Name}
	m.data[1] = Status{Key: "Status", Value: status.State}
	m.data[2] = Status{Key: "Domain", Value: status.Domain}
	m.data[3] = Status{Key: "Virtual IP", Value: status.VirtualIP}
	m.data[4] = Status{Key: "Controller Address", Value: status.ControllerAddress}
	m.data[5] = Status{Key: "My Public Address", Value: status.PublicAddress}
	m.data[6] = Status{Key: "My Public Key", Value: status.PublicKey}

	return m
}

func (m *StatusModel) RowCount() int {
	m.RLock()
	defer m.RUnlock()

	return len(m.data)
}

func (m *StatusModel) Value(row int, col int) interface{} {
	m.RLock()
	defer m.RUnlock()

	item := m.data[row]

	switch col {
	case 0:
		return item.Key

	case 1:
		return item.Value
	}

	panic("unexpected col")
}

func newStatusPage(parent walk.Container) (Page, error) {
	p := new(StatusPage)
	model := NewStatusModel()

	if err := (Composite{
		Background: SolidColorBrush{Color: BACKGROUND_COLOR},
		AssignTo:   &p.Composite,
		Name:       "StatusPage",
		Layout:     VBox{},
		Children: []Widget{
			Composite{
				Layout:  HBox{},
				MaxSize: Size{Width: 500, Height: 295},
				Children: []Widget{
					TableView{
						Background:          SolidColorBrush{Color: BACKGROUND_TABLE},
						AssignTo:            &p.tbView,
						LastColumnStretched: true,                      // Stretch the last column to fill the available space
						StretchFactor:       1,                         // Stretch the TableView itself to fill the available space
						MaxSize:             Size{Width: 0, Height: 0}, // Set the height to the maximum visible area without scrolling
						Columns: []TableViewColumn{
							{Title: "", Width: 150},
							{Title: "", Width: 400},
						},
						Model: model,
					},
				},
			},
			Composite{
				Background:         SolidColorBrush{Color: BACKGROUND_COLOR},
				Layout:             HBox{},
				MaxSize:            Size{Width: 500, Height: 100},
				AlwaysConsumeSpace: true,
				Children: []Widget{
					PushButton{
						Background: SolidColorBrush{Color: BACKGROUND_COLOR},
						MaxSize:    Size{Width: 100, Height: 50},
						Alignment:  AlignHNearVNear,
						Text:       "Terminate",
						OnClicked:  onQuit,
					},
				},
			},
		},
	}).Create(NewBuilder(parent)); err != nil {
		return nil, err
	}

	if err := walk.InitWrapperWindow(p); err != nil {
		return nil, err
	}
	p.tbView.SetHeaderHidden(true)
	b, _ := walk.NewSolidColorBrush(BACKGROUND_COLOR)
	p.tbView.SetBackground(b)
	p.tbView.SetCellStyler(p)
	go updateModel(model, p)

	return p, nil
}

// StyleCell sets the background color for each cell based on the row index
func (s *StatusPage) StyleCell(style *walk.CellStyle) {
	style.BackgroundColor = BACKGROUND_TABLE
	style.TextColor = TEXT_WHITE
}

func updateModel(m *StatusModel, p *StatusPage) {
	for range eclipz.ClientStatusNotification {
		model := NewStatusModel()
		if model != nil {
			p.tbView.SetModel(model)
		}
	}
}

func (mw *AgentWindow) buttonHandler1() {
	showStatusPage(mw)
}

func showStatusPage(mw *AgentWindow) {
	if prevPage := mw.currentPage; prevPage != nil {
		mw.pageCom.SaveState()
		prevPage.SetVisible(false)
		prevPage.(walk.Widget).SetParent(nil)
		prevPage.Dispose()
	}
	page, err := newStatusPage(mw.pageCom)
	if err != nil {
		return
	}
	mw.currentPage = page
}

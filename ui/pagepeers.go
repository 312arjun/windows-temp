/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"golang.zx2c4.com/wireguard/windows/eclipz"
)

type PeersPage struct {
	*walk.Composite

	tlStatus  *walk.TextLabel
	comPeers  *walk.Composite
	peers     []*eclipz.ApiPeer
	firstPeer string
	tbView    *walk.TableView
}

type PeerData struct {
	Name  string
	Value string
}

type PeerModel struct {
	walk.TableModelBase
	items []*PeerData
}

func (m *PeerModel) RowCount() int {
	return len(m.items)
}

func (m *PeerModel) Value(row, col int) interface{} {
	item := m.items[row]
	if item == nil {
		return ""
	}

	switch col {
	case 0:
		return item.Name
	case 1:
		return item.Value
	}

	panic("unexpected col")
}

func getPeers() ([]*eclipz.ApiPeer, error) {
	//// Get status from API
	url := "http://localhost:8089/api/peers"
	resp, err := GET(url, "", nil)
	if err != nil {
		fmt.Printf("Get peers: %v\n", err)
		return nil, err
	}

	var peers []*eclipz.ApiPeer
	err = json.Unmarshal(resp, &peers)
	if err != nil {
		fmt.Printf("Peers: Unmarshal %v\n", err)
		return nil, err
	}

	// Get peers from eclipz module
	// peers := eclipz.GetPeers()

	return peers, nil
}

func newPeersPage(parent walk.Container) (Page, error) {
	p := new(PeersPage)

	if err := (Composite{
		AssignTo:   &p.Composite,
		Alignment:  AlignHNearVNear,
		Name:       "PeersPage",
		Layout:     VBox{},
		Background: SolidColorBrush{Color: BACKGROUND_COLOR},
		Children: []Widget{
			Composite{
				Layout:  HBox{},
				MinSize: Size{Width: 500, Height: 293},
				Children: []Widget{
					Composite{
						Layout:     HBox{MarginsZero: true},
						Background: SolidColorBrush{Color: BACKGROUND_COLOR},
						Children: []Widget{
							ScrollView{
								Background:      SolidColorBrush{Color: BACKGROUND_COLOR},
								HorizontalFixed: true,
								Layout:          VBox{MarginsZero: true},
								MaxSize:         Size{Width: 132, Height: 293}, // Set the height to the maximum visible area without scrolling
								Children: []Widget{
									Composite{
										AssignTo: &p.comPeers,
										Border:   true,
										Layout:   VBox{},
										MaxSize:  Size{Width: 135, Height: 0}, // Set the height to the maximum visible area without scrolling
									},
								},
							},
							Composite{
								Layout: VBox{MarginsZero: true},
								Children: []Widget{
									TableView{
										AssignTo:            &p.tbView,
										LastColumnStretched: true,                          // Stretch the last column to fill the available space
										StretchFactor:       1,                             // Stretch the TableView itself to fill the available space
										MaxSize:             Size{Width: 125, Height: 293}, // Set the height to the maximum visible area without scrolling
										Columns: []TableViewColumn{
											{Title: "", Width: 150},
											{Title: "", Width: 300},
										},
									},
								},
							},
						},
					},
				},
			},
			Composite{
				Layout:  HBox{},
				MaxSize: Size{Width: 200, Height: 100},
				Font:    Font{Family: "Segoe UI", PointSize: 12},
				Children: []Widget{
					TextLabel{
						TextColor: TEXT_WHITE,
						AssignTo:  &p.tlStatus,
						MaxSize:   Size{Width: 125, Height: 50},
						Alignment: AlignHNearVNear,
					},
					PushButton{
						MaxSize:   Size{Width: 100},
						Text:      "Disconnect",
						Alignment: AlignHNearVNear,
						OnClicked: onQuit,
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

	p.tlStatus.SetText("Connected")
	wgFont, _ := walk.NewFont("Segoe UI", 10, 0)
	p.SetFont(wgFont)

	// Add as many buttons as number of peers
	// peers, _ := getPeers()
	peers := eclipz.GetPeers()
	p.peers = peers
	for i, peer := range peers {
		button, _ := walk.NewPushButton(p.comPeers)
		button.SetText(peer.Name)
		button.SetAlignment(walk.AlignHCenterVNear)
		button.SetFont(wgFont)
		handler := p.getPeerHandler(peer.Name)
		button.Clicked().Attach(handler)

		if i == 0 {
			p.firstPeer = peer.Name
		}
	}
	p.tbView.SetHeaderHidden(true)
	b, _ := walk.NewSolidColorBrush(BACKGROUND_COLOR)
	p.tbView.SetBackground(b)
	p.tbView.SetCellStyler(p)

	p.getPeerHandler(p.firstPeer)()

	return p, nil
}

// StyleCell sets the background color for each cell based on the row index
func (s *PeersPage) StyleCell(style *walk.CellStyle) {
	style.BackgroundColor = BACKGROUND_TABLE
	style.TextColor = TEXT_WHITE
}

func (p *PeersPage) getPeerHandler(name string) func() {
	return func() {
		for _, peer2 := range p.peers {
			if peer2.Name == name {
				model := new(PeerModel)

				model.items = make([]*PeerData, 13)

				model.items[0] = &PeerData{Name: "Endpoint", Value: peer2.Endpoint}
				model.items[1] = &PeerData{Name: "Virtual IP", Value: peer2.VirtualIP}
				model.items[2] = &PeerData{Name: "Uptime", Value: peer2.Uptime}
				model.items[3] = &PeerData{Name: "Public Key", Value: peer2.WgKey}
				model.items[4] = &PeerData{Name: "Applications\nAllowed", Value: peer2.AllowedIPs}
				model.items[5] = &PeerData{Name: "Rx Bytes", Value: strconv.FormatInt(peer2.RxBytes, 10)}
				model.items[6] = &PeerData{Name: "Tx Bytes", Value: strconv.FormatInt(peer2.TxBytes, 10)}
				model.items[7] = &PeerData{Name: "Last\nHandshake", Value: strconv.Itoa(peer2.LastHandshakeSecs)}

				//Set model
				p.tbView.SetModel(model)
				//p.tbView.SetGridlines(true)
				break
			}
		}
	}
}

func (mw *AgentWindow) buttonHandler3() {
	if prevPage := mw.currentPage; prevPage != nil {
		mw.pageCom.SaveState()
		prevPage.SetVisible(false)
		prevPage.(walk.Widget).SetParent(nil)
		prevPage.Dispose()
	}

	page, err := newPeersPage(mw.pageCom)
	if err != nil {
		return
	}

	mw.currentPage = page
}

/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"errors"
	"os"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/windows/eclipz"
	"golang.zx2c4.com/wireguard/windows/l18n"
)

type Page interface {
	walk.Container
	Parent() walk.Container
	SetParent(parent walk.Container) error
}

type AgentWindow struct {
	*walk.MainWindow

	logo        *walk.ImageView
	pageCom     *walk.Composite
	currentPage Page
}

var BACKGROUND_COLOR = walk.RGB(12, 19, 26)
var BACKGROUND_TABLE = walk.RGB(21, 33, 45)
var TEXT_WHITE = walk.RGB(255, 255, 255)

var eclipzHomeDir = "C:\\Program Files\\Eclipz"

func NewAgentWindow() (*AgentWindow, error) {
	var disposables walk.Disposables
	defer disposables.Treat()

	mw := new(AgentWindow)

	// For loading images
	walk.Resources.SetRootDirPath(eclipzHomeDir)
	var imageFile string = eclipzHomeDir + "\\img\\01.bmp"
	if _, err := os.Stat(imageFile); errors.Is(err, os.ErrNotExist) {
		return nil, errors.New("no images for menu")
	}

	if err := (MainWindow{
		AssignTo:   &mw.MainWindow,
		Title:      "Eclipz Agent",
		Font:       Font{Family: "Segoe UI", PointSize: 12},
		Layout:     Grid{Columns: 2},
		Background: SolidColorBrush{Color: BACKGROUND_COLOR},
		Children: []Widget{
			Composite{
				Border:     true,
				Alignment:  AlignHNearVNear,
				Background: SolidColorBrush{Color: BACKGROUND_COLOR},
				Layout:     VBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 160}},
				Children: []Widget{
					ImageView{
						Image:   "img/01.bmp",
						MaxSize: Size{Width: 96, Height: 35},
						OnMouseUp: func(x, y int, button walk.MouseButton) {
							mw.buttonHandler1()
						},
					},
					ImageView{
						Image:   "img/02.bmp",
						MaxSize: Size{Width: 96, Height: 35},
						OnMouseUp: func(x, y int, button walk.MouseButton) {
							mw.buttonHandler2()
						},
					},
					ImageView{
						Image:   "img/03.bmp",
						MaxSize: Size{Width: 96, Height: 35},
						OnMouseUp: func(x, y int, button walk.MouseButton) {
							mw.buttonHandler3()
						},
					},
					ImageView{
						Image:   "img/05.bmp",
						MaxSize: Size{Width: 96, Height: 35},
						OnMouseUp: func(x, y int, button walk.MouseButton) {
							newResetConfigWindow(mw)
						},
					},
					ImageView{
						Image:   "img/04.bmp",
						MaxSize: Size{Width: 96, Height: 35},
						OnMouseUp: func(x, y int, button walk.MouseButton) {
							mw.buttonHandler4()
						},
					},
					VSpacer{Size: 30},
					ImageView{
						AssignTo: &mw.logo,
					},
				},
			},
			Composite{
				Alignment: AlignHNearVNear,
				Border:    true,
				AssignTo:  &mw.pageCom,
				Name:      "pageCom",
				Layout:    HBox{MarginsZero: true},
			},
		},
	}.Create()); err != nil {
		return nil, err
	}

	// Set the logo
	if logo, err := loadLogoIconWhite(48); err == nil {
		mw.logo.SetImage(logo)
	}
	mw.logo.Accessibility().SetName(l18n.Sprintf("Eclipz logo image"))

	// Set behavior when closing
	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		// "Close to tray" instead of exiting application
		*canceled = true
		if !noTrayAvailable {
			mw.Hide()
		} else {
			win.ShowWindow(mw.Handle(), win.SW_MINIMIZE)
		}
	})

	go func() {
		for {
			select {
			case msg := <-eclipz.ErrorChannel:
				newErrorMessage(mw.MainWindow, msg)
			case <-eclipz.RequestCredentials:
				NewManualConfig(mw.MainWindow)
			}
		}
	}()

	// Change the window style so not resizable
	win.SetWindowLong(mw.Handle(), win.GWL_STYLE, win.GetWindowLong(mw.Handle(), win.GWL_STYLE) & ^win.WS_MAXIMIZEBOX & ^win.WS_SIZEBOX)

	// Set the status page active
	mw.buttonHandler1()

	disposables.Spare()

	return mw, nil
}

func (mw *AgentWindow) Dispose() {

	mw.FormBase.Dispose()
}

// func (mw *AgentWindow) UpdateFound() {
// 	if mw.updatePage != nil {
// 		return
// 	}
// 	if IsAdmin {
// 		mw.SetTitle(l18n.Sprintf("%s (out of date)", mw.Title()))
// 	}
// 	updatePage, err := NewUpdatePage()
// 	if err == nil {
// 		mw.updatePage = updatePage
// 		mw.tabs.Pages().Add(updatePage.TabPage)
// 	}
// }

func (mw *AgentWindow) buttonHandler5() {
	if prevPage := mw.currentPage; prevPage != nil {
		mw.pageCom.SaveState()
		prevPage.SetVisible(false)
		prevPage.(walk.Widget).SetParent(nil)
		prevPage.Dispose()
	}

	page, err := NewUpdatePage2(mw.pageCom)
	if err != nil {
		return
	}

	mw.currentPage = page
}

func (mw *AgentWindow) WndProc(hwnd win.HWND, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case win.WM_QUERYENDSESSION:
		if lParam == win.ENDSESSION_CLOSEAPP {
			return win.TRUE
		}
	case win.WM_ENDSESSION:
		if lParam == win.ENDSESSION_CLOSEAPP && wParam == 1 {
			walk.App().Exit(198)
		}
	case win.WM_SYSCOMMAND:
		if wParam == aboutWireGuardCmd {
			// onAbout(mtw)
			return 0
		}
	case raiseMsg:
		windows.MessageBox(0, windows.StringToUTF16Ptr("Raise Message"), windows.StringToUTF16Ptr("Error"), windows.MB_ICONERROR)

		// if mtw.tunnelsPage == nil || mtw.tabs == nil {
		// 	mtw.Synchronize(func() {
		// 		mtw.SendMessage(msg, wParam, lParam)
		// 	})
		// 	return 0
		// }
		// if !mtw.Visible() {
		// 	mtw.tunnelsPage.listView.SelectFirstActiveTunnel()
		// 	if mtw.tabs.Pages().Len() != 3 {
		// 		mtw.tabs.SetCurrentIndex(0)
		// 	}
		// }
		// if mtw.tabs.Pages().Len() == 3 {
		// 	mtw.tabs.SetCurrentIndex(2)
		// }
		// raise(mtw.Handle())
		return 0
		// case taskbarButtonCreatedMsg:
		// 	ret := mtw.FormBase.WndProc(hwnd, msg, wParam, lParam)
		// 	go func() {
		// 		globalState, err := manager.IPCClientGlobalState()
		// 		if err == nil {
		// 			mtw.Synchronize(func() {
		// 				mtw.updateProgressIndicator(globalState)
		// 			})
		// 		}
		// 	}()
		// 	return ret
	}

	return mw.FormBase.WndProc(hwnd, msg, wParam, lParam)
}

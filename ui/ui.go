/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/lxn/walk"
	"golang.org/x/sys/windows"

	"golang.zx2c4.com/wireguard/windows/l18n"
	"golang.zx2c4.com/wireguard/windows/manager"
	"golang.zx2c4.com/wireguard/windows/tunnel"
	"golang.zx2c4.com/wireguard/windows/version"
)

var (
	noTrayAvailable              = false
	shouldQuitManagerWhenExiting = false
	startTime                    = time.Now()
	IsAdmin                      = false // A global, because this really is global for the process
)

func RunUI() {
	runtime.LockOSThread()
	windows.SetProcessPriorityBoost(windows.CurrentProcess(), false)
	defer func() {
		if err := recover(); err != nil {
			showErrorCustom(nil, "Panic", fmt.Sprint(err, "\n\n", string(debug.Stack())))
			panic(err)
		}
	}()

	var (
		err  error
		atw  *AgentWindow
		tray *Tray2
	)

	// .........Not sure why such loop needs
	// for atw == nil {
	// 	atw, err = NewAgentWindow()
	// 	if err != nil {
	// 		tunnel.WGlog("ui: Error while creating the main window %+v", err)
	// 		time.Sleep(time.Millisecond * 400)
	// 	}
	// }
	atw, err = NewAgentWindow()
	if err != nil {
		tunnel.WGlog("ui: Error while creating the main window %+v", err)
		onQuit()
	}

	for tray == nil {
		tray, err = NewTray2(atw)
		if err != nil {
			if version.OsIsCore() {
				noTrayAvailable = true
				break
			}
			time.Sleep(time.Millisecond * 400)
		}
	}

	manager.IPCClientRegisterManagerStopping(func() {
		atw.Synchronize(func() {
			walk.App().Exit(0)
		})
	})

	onUpdateNotification := func(updateState manager.UpdateState) {
		if updateState == manager.UpdateStateUnknown {
			return
		}
		atw.Synchronize(func() {
			switch updateState {
			case manager.UpdateStateFoundUpdate:
				atw.buttonHandler5()
				if tray != nil && IsAdmin {
					tray.UpdateFound()
				}
			case manager.UpdateStateUpdatesDisabledUnofficialBuild:
				atw.SetTitle(l18n.Sprintf("%s (unsigned build, no updates)", atw.Title()))
			}
		})
	}
	manager.IPCClientRegisterUpdateFound(onUpdateNotification)
	go func() {
		updateState, err := manager.IPCClientUpdateState()
		if err == nil {
			onUpdateNotification(updateState)
		}
	}()

	tray.clicked()
	atw.SetSize(walk.Size{Width: 800, Height: 400})

	atw.Run()
	if tray != nil {
		tray.Dispose()
	}
	atw.Dispose()

	if shouldQuitManagerWhenExiting {
		_, err := manager.IPCClientQuit(true)
		if err != nil {
			showErrorCustom(nil, l18n.Sprintf("Error Exiting Eclipz"), l18n.Sprintf("Unable to exit service due to: %v. You may want to stop Eclipz from the service manager.", err))
		}
	}
}

func onQuit() {
	shouldQuitManagerWhenExiting = true
	walk.App().Exit(0)
}

func showError(err error, owner walk.Form) bool {
	if err == nil {
		return false
	}

	showErrorCustom(owner, l18n.Sprintf("Error"), err.Error())

	return true
}

func showErrorCustom(owner walk.Form, title, message string) {
	walk.MsgBox(owner, title, message, walk.MsgBoxIconError)
}

func showWarningCustom(owner walk.Form, title, message string) {
	walk.MsgBox(owner, title, message, walk.MsgBoxIconWarning)
}

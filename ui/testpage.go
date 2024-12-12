/* SPDX-License-Identifier: MIT
 *
 * Copyright (C) 2019-2022 WireGuard LLC. All Rights Reserved.
 */

package ui

import (
	"github.com/lxn/walk"
	"golang.zx2c4.com/wireguard/windows/l18n"
)

type TestPage struct {
	*walk.TabPage
	testView *walk.TableView
}

func NewTestPage() (*TestPage, error) {
	lp := &TestPage{}

	var err error
	var disposables walk.Disposables
	defer disposables.Treat()

	if lp.TabPage, err = walk.NewTabPage(); err != nil {
		return nil, err
	}
	disposables.Add(lp)

	lp.SetTitle(l18n.Sprintf("Test"))
	lp.SetLayout(walk.NewVBoxLayout())

	if lp.testView, err = walk.NewTableView(lp); err != nil {
		return nil, err
	}
	lp.testView.SetAlternatingRowBG(true)
	lp.testView.SetLastColumnStretched(true)
	lp.testView.SetGridlines(true)

	nameCol := walk.NewTableViewColumn()
	nameCol.SetName("Name")
	nameCol.SetTitle(l18n.Sprintf("Friendly Name"))
	nameCol.SetWidth(140)
	lp.testView.Columns().Add(nameCol)

	addressCol := walk.NewTableViewColumn()
	addressCol.SetName("Address")
	addressCol.SetTitle(l18n.Sprintf("Physical Address"))
	lp.testView.Columns().Add(addressCol)

	disposables.Spare()

	return lp, nil
}

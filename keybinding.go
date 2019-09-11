package main

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/blrsn/zentile/state"
	log "github.com/sirupsen/logrus"
)

type keyMapper struct{}

func (k keyMapper) bind(action string, f func()) {
	err := keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			f()
		}).Connect(state.X, state.X.RootWin(),
		Config.Keybindings[action], true)

	if err != nil {
		log.Warn(err)
	}
}

func bindKeys(t *tracker) {
	workspaces := t.workspaces
	keybind.Initialize(state.X)
	k := keyMapper{}

	k.bind("tile", func() {
		ws := workspaces[state.CurrentDesk]
		ws.IsTiling = true
		ws.Tile()
	})
	k.bind("untile", func() {
		ws := workspaces[state.CurrentDesk]
		ws.Untile()
	})
	k.bind("make_active_window_master", func() {
		c := t.clients[state.ActiveWin]
		ws := workspaces[state.CurrentDesk]
		ws.ActiveLayout().MakeMaster(c)
		ws.Tile()
	})
	k.bind("switch_layout", func() {
		workspaces[state.CurrentDesk].SwitchLayout()
	})
	k.bind("increase_master", func() {
		ws := workspaces[state.CurrentDesk]
		ws.ActiveLayout().IncMaster()
		ws.Tile()
	})
	k.bind("decrease_master", func() {
		ws := workspaces[state.CurrentDesk]
		ws.ActiveLayout().DecreaseMaster()
		ws.Tile()
	})
	k.bind("next_window", func() {
		ws := workspaces[state.CurrentDesk]
		ws.ActiveLayout().NextClient()
	})
	k.bind("previous_window", func() {
		ws := workspaces[state.CurrentDesk]
		ws.ActiveLayout().PreviousClient()
	})
	k.bind("increment_master", func() {
		ws := workspaces[state.CurrentDesk]
		ws.ActiveLayout().IncrementMaster()
		ws.Tile()
	})
	k.bind("decrement_master", func() {
		ws := workspaces[state.CurrentDesk]
		ws.ActiveLayout().DecrementMaster()
		ws.Tile()
	})
}

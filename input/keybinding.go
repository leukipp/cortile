package input

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"

	log "github.com/sirupsen/logrus"
)

type KeyMapper struct {
}

func BindKeys(t *desktop.Tracker) {
	workspaces := t.Workspaces
	keybind.Initialize(common.X)
	k := KeyMapper{}

	k.bind("tile", func() {
		ws := workspaces[common.CurrentDesk]
		ws.IsTiling = true
		ws.Tile()
	})
	k.bind("untile", func() {
		ws := workspaces[common.CurrentDesk]
		ws.UnTile()
	})
	k.bind("make_active_window_master", func() {
		c := t.Clients[common.ActiveWin]
		ws := workspaces[common.CurrentDesk]
		ws.ActiveLayout().MakeMaster(c)
		ws.Tile()
	})
	k.bind("switch_layout", func() {
		workspaces[common.CurrentDesk].SwitchLayout()
	})
	k.bind("increase_master", func() {
		ws := workspaces[common.CurrentDesk]
		ws.ActiveLayout().IncreaseMaster()
		ws.Tile()
	})
	k.bind("decrease_master", func() {
		ws := workspaces[common.CurrentDesk]
		ws.ActiveLayout().DecreaseMaster()
		ws.Tile()
	})
	k.bind("next_window", func() {
		ws := workspaces[common.CurrentDesk]
		ws.ActiveLayout().NextClient()
	})
	k.bind("previous_window", func() {
		ws := workspaces[common.CurrentDesk]
		ws.ActiveLayout().PreviousClient()
	})
	k.bind("increment_master", func() {
		ws := workspaces[common.CurrentDesk]
		ws.ActiveLayout().IncrementProportion()
		ws.Tile()
	})
	k.bind("decrement_master", func() {
		ws := workspaces[common.CurrentDesk]
		ws.ActiveLayout().DecrementProportion()
		ws.Tile()
	})
}

func (k KeyMapper) bind(action string, f func()) {
	err := keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			// callback
			f()
		}).Connect(common.X, common.X.RootWin(), common.Config.Keybindings[action], true)

	if err != nil {
		log.Warn(err)
	}
}

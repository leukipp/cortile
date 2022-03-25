package input

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"

	log "github.com/sirupsen/logrus"
)

type KeyMapper struct{}

func BindKeys(t *desktop.Tracker) {
	keybind.Initialize(common.X)
	k := KeyMapper{}

	// Bind keyboard shortcuts
	k.bind("tile", func() {
		ws := t.Workspaces[common.CurrentDesk]
		ws.TilingEnabled = true
		ws.Tile()
	})
	k.bind("untile", func() {
		ws := t.Workspaces[common.CurrentDesk]
		ws.TilingEnabled = false
		ws.UnTile()
	})
	k.bind("make_active_window_master", func() {
		c := t.Clients[common.ActiveWin]
		ws := t.Workspaces[common.CurrentDesk]
		if !ws.TilingEnabled {
			return
		}
		ws.ActiveLayout().MakeMaster(c)
		ws.Tile()
	})
	k.bind("switch_layout", func() {
		ws := t.Workspaces[common.CurrentDesk]
		if !ws.TilingEnabled {
			return
		}
		ws.SwitchLayout()
	})
	k.bind("increase_master", func() {
		ws := t.Workspaces[common.CurrentDesk]
		if !ws.TilingEnabled {
			return
		}
		ws.ActiveLayout().IncreaseMaster()
		ws.Tile()
	})
	k.bind("decrease_master", func() {
		ws := t.Workspaces[common.CurrentDesk]
		if !ws.TilingEnabled {
			return
		}
		ws.ActiveLayout().DecreaseMaster()
		ws.Tile()
	})
	k.bind("next_window", func() {
		ws := t.Workspaces[common.CurrentDesk]
		if !ws.TilingEnabled {
			return
		}
		ws.ActiveLayout().NextClient()
	})
	k.bind("previous_window", func() {
		ws := t.Workspaces[common.CurrentDesk]
		if !ws.TilingEnabled {
			return
		}
		ws.ActiveLayout().PreviousClient()
	})
	k.bind("increment_master", func() {
		ws := t.Workspaces[common.CurrentDesk]
		if !ws.TilingEnabled {
			return
		}
		ws.ActiveLayout().IncrementProportion()
		ws.Tile()
	})
	k.bind("decrement_master", func() {
		ws := t.Workspaces[common.CurrentDesk]
		if !ws.TilingEnabled {
			return
		}
		ws.ActiveLayout().DecrementProportion()
		ws.Tile()
	})
}

func (k KeyMapper) bind(action string, f func()) {
	err := keybind.KeyPressFun(
		func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
			// Callback
			f()
		}).Connect(common.X, common.X.RootWin(), common.Config.Keys[action], true)

	if err != nil {
		log.Warn(err)
	}
}

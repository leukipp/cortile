package input

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type KeyMapper struct{}

func BindKeys(tr *desktop.Tracker) {
	keybind.Initialize(store.X)

	// Bind keyboard shortcuts
	k := KeyMapper{}
	k.bind("tile", tr)
	k.bind("untile", tr)
	k.bind("layout_cycle", tr)
	k.bind("layout_fullscreen", tr)
	k.bind("layout_vertical_left", tr)
	k.bind("layout_vertical_right", tr)
	k.bind("layout_horizontal_top", tr)
	k.bind("layout_horizontal_bottom", tr)
	k.bind("master_make", tr)
	k.bind("master_increase", tr)
	k.bind("master_decrease", tr)
	k.bind("slave_increase", tr)
	k.bind("slave_decrease", tr)
	k.bind("proportion_increase", tr)
	k.bind("proportion_decrease", tr)
	k.bind("window_next", tr)
	k.bind("window_previous", tr)
}

func (k KeyMapper) bind(a string, tr *desktop.Tracker) {
	err := keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		Execute(a, tr)
	}).Connect(store.X, store.X.RootWin(), common.Config.Keys[a], true)
	if err != nil {
		log.Warn("Error on action for ", a, ": ", err)
	}
}

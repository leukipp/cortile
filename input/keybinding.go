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

	// Bind keyboard shortcuts
	k := KeyMapper{}
	k.bind("tile", t)
	k.bind("untile", t)
	k.bind("layout_cycle", t)
	k.bind("layout_fullscreen", t)
	k.bind("layout_vertical_left", t)
	k.bind("layout_vertical_right", t)
	k.bind("layout_horizontal_top", t)
	k.bind("layout_horizontal_bottom", t)
	k.bind("master_make", t)
	k.bind("master_increase", t)
	k.bind("master_decrease", t)
	k.bind("slave_increase", t)
	k.bind("slave_decrease", t)
	k.bind("proportion_increase", t)
	k.bind("proportion_decrease", t)
	k.bind("window_next", t)
	k.bind("window_previous", t)
}

func (k KeyMapper) bind(a string, t *desktop.Tracker) {
	err := keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		Execute(a, t)
	}).Connect(common.X, common.X.RootWin(), common.Config.Keys[a], true)
	if err != nil {
		log.Warn("Error on action for ", a, ": ", err)
	}
}

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

func BindKeys(tr *desktop.Tracker) {
	keybind.Initialize(store.X)

	// Bind keyboard shortcuts
	for a, k := range common.Config.Keys {
		bind(a, k, tr)
	}
}

func bind(action string, key string, tr *desktop.Tracker) {
	err := keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		Execute(action, tr)
	}).Connect(store.X, store.X.RootWin(), key, true)
	if err != nil {
		log.Warn("Error on action for ", action, ": ", err)
	}
}

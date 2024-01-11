package input

import (
	"strings"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/desktop"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

func BindKeys(tr *desktop.Tracker) {
	keybind.Initialize(store.X)

	actions := map[string]string{}
	mods := map[string]string{"current": ""}

	// Map actions and modifiers
	for c, ck := range common.Config.Keys {
		if !strings.HasPrefix(c, "mod_") {
			actions[c] = ck
		} else {
			mods[c[4:]] = ck
		}
	}

	// Bind keyboard shortcuts
	for a, ak := range actions {
		for m, mk := range mods {
			if len(mk) == 0 {
				bind(ak, a, m, tr)
			} else {
				bind(mk+"-"+ak, a, m, tr)
			}
		}
	}
}

func bind(key string, action string, mod string, tr *desktop.Tracker) {
	err := keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		Execute(action, mod, tr)
	}).Connect(store.X, store.X.RootWin(), key, true)

	if err != nil {
		log.Warn("Error on action for ", action, ": ", err)
	}
}

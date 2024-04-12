package input

import (
	"strings"

	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/keybind"
	"github.com/jezek/xgbutil/xevent"

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

	// Bind action channel
	go action(tr.Channels.Action, tr)
}

func bind(key string, action string, mod string, tr *desktop.Tracker) {
	err := keybind.KeyPressFun(func(X *xgbutil.XUtil, ev xevent.KeyPressEvent) {
		ExecuteActions(action, tr, mod)
	}).Connect(store.X, store.X.RootWin(), key, true)

	if err != nil {
		log.Warn("Error on action ", action, ": ", err)
	}
}

func action(ch chan string, tr *desktop.Tracker) {
	for {
		ExecuteAction(<-ch, tr, tr.ActiveWorkspace())
	}
}

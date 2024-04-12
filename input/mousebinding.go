package input

import (
	"time"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/desktop"
	"github.com/leukipp/cortile/v2/store"
	"github.com/leukipp/cortile/v2/ui"

	log "github.com/sirupsen/logrus"
)

var (
	workspace *desktop.Workspace // Stores last active workspace (for comparison only)
)

func BindMouse(tr *desktop.Tracker) {
	poll(100, func() {
		store.PointerUpdate(store.X)

		// Compare active workspace
		ws := tr.ActiveWorkspace()
		if ws != workspace {
			log.Info("Active workspace changed [", ws.Name, "]")

			// Communicate workplace change
			tr.Channels.Event <- "workplace_change"

			// Update systray icon
			ui.UpdateIcon(ws)

			// Store last workspace
			workspace = ws
		}

		// Evaluate corner states
		hc := store.HotCorner()
		if hc != nil {

			// Communicate corner change
			tr.Channels.Event <- "corner_change"

			// Execute action
			ExecuteAction(common.Config.Corners[hc.Name], tr, tr.ActiveWorkspace())
		}
	})
}

func poll(t time.Duration, fun func()) {
	go func() {
		for range time.Tick(t * time.Millisecond) {
			fun()
		}
	}()
}

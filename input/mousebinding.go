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
	workspace *desktop.Workspace // Stores last active workspace
)

func BindMouse(tr *desktop.Tracker) {
	poll(100, func() {
		pt := store.PointerUpdate(store.X)

		// Update systray icon
		ws := tr.ActiveWorkspace()
		if ws != workspace {
			ui.UpdateIcon(ws)
			workspace = ws
		}

		// Evaluate corner states
		for i := range store.Corners {
			hc := store.Corners[i]

			wasActive := hc.Active
			isActive := hc.IsActive(pt)

			if !wasActive && isActive {
				log.Debug("Corner at position ", hc.Area, " is hot [", hc.Name, "]")
				Execute(common.Config.Corners[hc.Name], "current", tr)
			} else if wasActive && !isActive {
				log.Debug("Corner at position ", hc.Area, " is cold [", hc.Name, "]")
			}
		}
	})
}

func poll(t time.Duration, fun func()) {
	fun()
	go func() {
		for range time.Tick(t * time.Millisecond) {
			_, err := store.X.Conn().PollForEvent()
			if err != nil {
				continue
			}
			fun()
		}
	}()
}

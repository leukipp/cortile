package input

import (
	"time"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"
	"github.com/leukipp/cortile/store"
	"github.com/leukipp/cortile/ui"

	log "github.com/sirupsen/logrus"
)

var (
	workspace *desktop.Workspace // Stores last active workspace
)

func BindMouse(tr *desktop.Tracker) {
	poll(50, func() {
		store.PointerUpdate(store.X)

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
			isActive := hc.IsActive(store.CurrentPointer)

			if !wasActive && isActive {
				log.Debug("Corner at position ", hc.Area, " is hot [", hc.Name, "]")
				Execute(common.Config.Corners[hc.Name], tr)
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

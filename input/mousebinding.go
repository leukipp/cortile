package input

import (
	"time"

	"github.com/BurntSushi/xgbutil"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

func BindMouse(tr *desktop.Tracker) {
	poll(store.X, 100, func() {
		store.PointerUpdate(store.X)

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

func poll(X *xgbutil.XUtil, t time.Duration, fun func()) {
	fun()
	go func() {
		for range time.Tick(t * time.Millisecond) {
			_, err := X.Conn().PollForEvent()
			if err != nil {
				continue
			}
			fun()
		}
	}()
}

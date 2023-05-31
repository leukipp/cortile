package input

import (
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"

	log "github.com/sirupsen/logrus"
)

func BindMouse(tr *desktop.Tracker) {
	poll(common.X, 50, func() {

		// Update pointer position
		common.Pointer, _ = xproto.QueryPointer(common.X.Conn(), common.X.RootWin()).Reply()

		// Update current screen
		common.CurrentScreen = common.ScreenNumGet(common.Pointer)

		// Evaluate corner states
		for i := range common.Corners {
			hc := common.Corners[i]

			wasActive := hc.Active
			isActive := hc.IsActive(common.Pointer)

			if !wasActive && isActive {
				log.Debug("Corner at position ", hc.Area, " is hot [", hc.Name, "]")
				Execute(common.Config.Corners[hc.Name], tr)
			} else if wasActive && !isActive {
				log.Debug("Corner at position ", hc.Area, " is cold [", hc.Name, "]")
			}
		}
	})
}

func poll(X *xgbutil.XUtil, t time.Duration, f func()) {
	go func() {
		for range time.Tick(t * time.Millisecond) {
			_, err := X.Conn().PollForEvent()
			if err != nil {
				continue
			}
			f()
		}
	}()
}

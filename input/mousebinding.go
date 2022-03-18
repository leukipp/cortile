package input

import (
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"

	log "github.com/sirupsen/logrus"
)

func BindMouse(t *desktop.Tracker) {
	backgroundTask(common.X, 100, func() {
		pointer, _ := xproto.QueryPointer(common.X.Conn(), common.X.RootWin()).Reply()
		x, y := pointer.RootX, pointer.RootY

		// Check corner states
		for i := range common.Corners {
			hc := &common.Corners[i]

			wasActive := hc.Active
			isActive := hc.IsActive(uint(x), uint(y))

			if !wasActive && isActive {
				log.Debug("Corner at position ", hc.Area, " is hot [", hc.Name, "]")

				// Get active clients and workspace
				c := t.Clients[common.ActiveWin]
				ws := t.Workspaces[common.CurrentDesk]

				// TODO: Load from config
				switch hc.Name {
				case "top_left":
					ws.SwitchLayout()
				case "top_center":
					// TODO: Add top-center
				case "top_right":
					ws.ActiveLayout().MakeMaster(c)
					ws.Tile()
				case "center_right":
					// TODO: Add center-right
				case "bottom_right":
					ws.ActiveLayout().IncreaseMaster()
					ws.Tile()
				case "bottom_center":
					// TODO: Add bottom-center
				case "bottom_left":
					ws.ActiveLayout().DecreaseMaster()
					ws.Tile()
				case "center_left":
					// TODO: Add center-left
				}
			} else if wasActive && !isActive {
				log.Debug("Corner at position ", hc.Area, " is cold [", hc.Name, "]")
			}
		}
	})
}

func backgroundTask(X *xgbutil.XUtil, t time.Duration, f func()) {
	go func() {
		// Poll X events in background
		for range time.Tick(t * time.Millisecond) {
			_, err := X.Conn().PollForEvent()
			if err != nil {
				continue
			}
			// Callback
			f()
		}
	}()
}

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
		// query mouse pointer
		pointer, _ := xproto.QueryPointer(common.X.Conn(), common.X.RootWin()).Reply()
		x, y := pointer.RootX, pointer.RootY

		//log.Trace("Pointer at ", "x=", x, ", y=", y)

		// update hotcorner states
		for i := range common.Corners {
			hc := &common.Corners[i]

			wasActive := hc.Active
			isActive := hc.IsActive(uint(x), uint(y))

			if !wasActive && isActive {
				// corner was entered
				log.Debug("Corner at position ", hc.Area, " is hot [", hc.Name, "]")

				// get active clients and workspace
				c := t.Clients[common.ActiveWin]
				ws := t.Workspaces[common.CurrentDesk]

				// TODO: load from config

				// execute hotcorner actions
				switch hc.Name {
				case "top_left":
					// switch_layout
					ws.SwitchLayout()
				case "top_center":
					// TODO: top center
				case "top_right":
					// make active window master
					ws.ActiveLayout().MakeMaster(c)
					ws.Tile()
				case "center_right":
					// TODO: center right
				case "bottom_right":
					// increase master
					ws.ActiveLayout().IncreaseMaster()
					ws.Tile()
				case "bottom_center":
					// TODO: bottom center
				case "bottom_left":
					// decrease master
					ws.ActiveLayout().DecreaseMaster()
					ws.Tile()
				case "center_left":
					// TODO: center left
				}
			} else if wasActive && !isActive {
				// corner was leaved
				log.Debug("Corner at position ", hc.Area, " is cold [", hc.Name, "]")
			}
		}
	})
}

// Poll X events in background
func backgroundTask(X *xgbutil.XUtil, t time.Duration, f func()) {
	go func() {
		for range time.Tick(time.Millisecond * t) {
			_, err := X.Conn().PollForEvent()
			if err != nil {
				continue
			}
			// callback
			f()
		}
	}()
}

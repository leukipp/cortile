package desktop

import (
	"math"
	"strings"
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type Tracker struct {
	Clients    map[xproto.Window]*store.Client // List of clients that are being tracked
	Workspaces map[uint]*Workspace             // List of workspaces used
}

func CreateTracker(ws map[uint]*Workspace) *Tracker {
	t := Tracker{
		Clients:    make(map[xproto.Window]*store.Client),
		Workspaces: ws,
	}

	// Init clients
	xevent.PropertyNotifyFun(t.handleClientUpdates).Connect(common.X, common.X.RootWin())
	t.populateClients()

	return &t
}

func (tr *Tracker) populateClients() {

	// Add trackable windows
	for _, w := range common.Stacking {
		if tr.isTrackable(w) {
			tr.trackWindow(w)
		}
	}

	// If window is tracked, but not in client list
	for w1 := range tr.Clients {
		trackable := false
		for _, w2 := range common.Stacking {
			if w1 == w2 {
				trackable = tr.isTrackable(w1)
				break
			}
		}
		if !trackable {
			tr.untrackWindow(w1)
		}
	}
}

func (tr *Tracker) trackWindow(w xproto.Window) {
	if tr.isTracked(w) {
		return
	}

	// Add new client
	c := store.CreateClient(w)
	tr.Clients[c.Win.Id] = &c
	ws := tr.Workspaces[c.Desk]
	ws.AddClient(&c)

	// Wait with handler attachment, as some applications load geometry delayed
	time.AfterFunc(1000*time.Millisecond, func() {
		tr.attachHandlers(&c)
		tr.Workspaces[common.CurrentDesk].Tile()
	})
}

func (tr *Tracker) untrackWindow(w xproto.Window) {
	if tr.isTracked(w) {
		c := tr.Clients[w]
		ws := tr.Workspaces[c.Desk]

		// Remove client
		ws.RemoveClient(c)
		xevent.Detach(common.X, w)
		delete(tr.Clients, w)
	}
}

func (tr *Tracker) handleResizeClient(c *store.Client) {

	// Previous dimensions
	pGeom := c.CurrentProp.Geom
	pw, ph := pGeom.Width(), pGeom.Height()

	// Current dimensions
	cGeom, err := c.Win.DecorGeometry()
	if err != nil {
		return
	}
	cw, ch := cGeom.Width(), cGeom.Height()

	// Update dimensions
	success := c.Update()
	if !success {
		return
	}

	// Re-tile on width or height change
	dw, dh := 0.0, 0.0
	tile := (math.Abs(float64(cw-pw)) > dw || math.Abs(float64(ch-ph)) > dh)

	// Tile workspace
	if tile {
		ws := tr.Workspaces[c.Desk]
		l := ws.ActiveLayout()
		s := l.GetManager()

		// Ignore master only windows
		if len(s.Slaves) == 0 {
			return
		}

		// Ignore fullscreen windows
		if store.IsMaximized(c.Win.Id) {
			return
		}

		proportion := 0.0
		gap := common.Config.WindowGap
		isMaster := ws.IsMaster(c)
		layoutType := l.GetType()
		_, _, dw, dh := common.DesktopDimensions()

		// Calculate proportion based on resized window width (TODO: LTR/RTL gap support)
		if layoutType == "vertical" {
			proportion = float64(cw+gap) / float64(dw)
			if isMaster {
				proportion = 1.0 - float64(cw+2*gap)/float64(dw)
			}
		}

		// Calculate proportion based on resized window height (TODO: LTR/RTL gap support)
		if layoutType == "horizontal" {
			proportion = 1.0 - float64(ch+gap)/float64(dh)
			if isMaster {
				proportion = float64(ch+2*gap) / float64(dh)
			}
		}

		log.Debug("Proportion set to ", proportion, " [", c.Class, "]")

		// Set proportion based on resized window
		l.SetProportion(proportion)
		ws.Tile()
	}
}

func (tr *Tracker) handleMaximizedClient(c *store.Client) {
	states, _ := ewmh.WmStateGet(common.X, c.Win.Id)

	// Client maximized
	for _, state := range states {
		if strings.Contains(state, "_NET_WM_STATE_MAXIMIZED") {
			ws := tr.Workspaces[c.Desk]
			for i, l := range ws.Layouts {
				if l.GetType() == "fullscreen" {
					ws.SetLayout(uint(i))
				}
			}
			ws.Tile()
		}
	}
}

func (tr *Tracker) handleMinimizedClient(c *store.Client) {
	states, _ := ewmh.WmStateGet(common.X, c.Win.Id)

	// Client minimized
	for _, state := range states {
		if state == "_NET_WM_STATE_HIDDEN" {
			tr.Workspaces[c.Desk].RemoveClient(c)
			tr.untrackWindow(c.Win.Id)
			tr.Workspaces[c.Desk].Tile()
		}
	}
}

func (tr *Tracker) handleDesktopChange(c *store.Client) {

	// Remove client from current workspace
	tr.Workspaces[c.Desk].RemoveClient(c)
	if tr.Workspaces[c.Desk].TilingEnabled {
		tr.Workspaces[c.Desk].Tile()
	}

	// Update client desktop
	success := c.Update()
	if !success {
		return
	}

	// Add client to new workspace
	tr.Workspaces[c.Desk].AddClient(c)
	if tr.Workspaces[c.Desk].TilingEnabled {
		tr.Workspaces[c.Desk].Tile()
	} else {
		c.Restore()
	}
}

func (tr *Tracker) handleClientUpdates(X *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
	aname, _ := xprop.AtomName(common.X, ev.Atom)

	// Client added or workspace changed
	if aname == "_NET_CLIENT_LIST_STACKING" || aname == "_NET_DESKTOP_VIEWPORT" {
		tr.populateClients()
		tr.Workspaces[common.CurrentDesk].Tile()
	}
}

func (tr *Tracker) attachHandlers(c *store.Client) {
	c.Win.Listen(xproto.EventMaskStructureNotify | xproto.EventMaskPropertyChange)

	// Attach resize events
	xevent.ConfigureNotifyFun(func(x *xgbutil.XUtil, ev xevent.ConfigureNotifyEvent) {
		log.Debug("Client configure event [", c.Class, "]")

		if tr.isTrackable(c.Win.Id) {
			tr.handleResizeClient(c)
		} else {
			tr.untrackWindow(c.Win.Id)
		}
	}).Connect(common.X, c.Win.Id)

	// Attach state events
	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		aname, _ := xprop.AtomName(common.X, ev.Atom)
		log.Debug("Client property event ", aname, " [", c.Class, "]")

		if tr.isTrackable(c.Win.Id) {
			if aname == "_NET_WM_STATE" {
				tr.handleMaximizedClient(c)
				tr.handleMinimizedClient(c)
			} else if aname == "_NET_WM_DESKTOP" {
				tr.handleDesktopChange(c)
			}
		} else {
			tr.untrackWindow(c.Win.Id)
		}
	}).Connect(common.X, c.Win.Id)
}

func (tr *Tracker) isTracked(w xproto.Window) bool {
	_, ok := tr.Clients[w]
	return ok
}

func (tr *Tracker) isTrackable(w xproto.Window) bool {
	return !store.IsHidden(w) && !store.IsModal(w) && !store.IsIgnored(w) && store.IsInsideViewPort(w)
}

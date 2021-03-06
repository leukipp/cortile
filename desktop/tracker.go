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
	tr.Clients[c.Win.Id] = c
	ws := tr.Workspaces[c.Info.Desk]
	ws.AddClient(c)

	// Attach handlers and tile
	tr.attachHandlers(c)
	tr.Workspaces[common.CurrentDesk].Tile()

	// Tile again as some applications load geometry delayed
	time.AfterFunc(1000*time.Millisecond, func() {
		tr.Workspaces[common.CurrentDesk].Tile()
	})
}

func (tr *Tracker) untrackWindow(w xproto.Window) {
	if tr.isTracked(w) {
		c := tr.Clients[w]
		ws := tr.Workspaces[c.Info.Desk]

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

	// Check width or height change
	resized := math.Abs(float64(cw-pw)) > 0.0 || math.Abs(float64(ch-ph)) > 0.0

	if resized {
		proportion := 0.0
		gap := common.Config.WindowGapSize
		ws := tr.Workspaces[c.Info.Desk]
		al := ws.ActiveLayout()
		mg := al.GetManager()

		// Update client dimensions
		success := c.Update()
		if !success {
			return
		}

		// Ignore fullscreen layouts
		if store.IsMaximized(c.Win.Id) {
			return
		}

		// Ignore master or slave only layouts
		if len(mg.Masters) == 0 || len(mg.Slaves) == 0 {
			return
		}

		// Calculate proportion based on resized window size
		_, _, dw, dh := common.DesktopDimensions()
		switch al.GetName() {
		case "vertical-left":
			proportion = 1.0 - (float64(cw+gap) / float64(dw))
			if mg.IsMaster(c) {
				proportion = float64(cw+2*gap) / float64(dw)
			}
		case "vertical-right":
			proportion = float64(cw+gap) / float64(dw)
			if mg.IsMaster(c) {
				proportion = 1.0 - (float64(cw+2*gap) / float64(dw))
			}
		case "horizontal-top":
			proportion = 1.0 - (float64(ch+gap) / float64(dh))
			if mg.IsMaster(c) {
				proportion = float64(ch+2*gap) / float64(dh)
			}
		case "horizontal-bottom":
			proportion = float64(ch+gap) / float64(dh)
			if mg.IsMaster(c) {
				proportion = 1.0 - (float64(ch+2*gap) / float64(dh))
			}
		}

		// Set proportion based on resized window
		log.Info("Proportion set to ", math.Round(proportion*100)/100, " [", c.Info.Class, "]")
		al.SetProportion(proportion)
		ws.Tile()
	}
}

func (tr *Tracker) handleMoveClient(c *store.Client) {

	// Previous position
	pGeom := c.CurrentProp.Geom
	px, py, _, _ := pGeom.Pieces()

	// Current position
	cGeom, err := c.Win.DecorGeometry()
	if err != nil {
		return
	}
	cx, cy, _, _ := cGeom.Pieces()

	// Check position change
	moved := math.Abs(float64(cx-px)) > 0.0 || math.Abs(float64(cy-py)) > 0.0

	if moved {
		ws := tr.Workspaces[c.Info.Desk]
		al := ws.ActiveLayout()
		mg := al.GetManager()

		// Check if pointer hovers other clients
		clients := mg.Clients()
		for _, co := range clients {
			if c.Win.Id == co.Win.Id {
				continue
			}

			// Update client dimensions
			success := co.Update()
			if !success {
				return
			}

			// Swap moved client with hovered client
			isHovered := common.IsInsideRect(common.Pointer, co.CurrentProp.Geom)
			if isHovered {
				log.Info("Swap clients [", c.Info.Class, " - ", co.Info.Class, "]")
				mg.SwapClient(c, co)
				break
			}
		}
		ws.Tile()
	}
}

func (tr *Tracker) handleMaximizedClient(c *store.Client) {
	states, _ := ewmh.WmStateGet(common.X, c.Win.Id)

	// Client maximized
	for _, state := range states {
		if strings.Contains(state, "_NET_WM_STATE_MAXIMIZED") {
			ws := tr.Workspaces[c.Info.Desk]
			for i, l := range ws.Layouts {
				if l.GetName() == "fullscreen" {
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
			tr.Workspaces[c.Info.Desk].RemoveClient(c)
			tr.untrackWindow(c.Win.Id)
			tr.Workspaces[c.Info.Desk].Tile()
		}
	}
}

func (tr *Tracker) handleDesktopChange(c *store.Client) {

	// Remove client from current workspace
	tr.Workspaces[c.Info.Desk].RemoveClient(c)
	if tr.Workspaces[c.Info.Desk].TilingEnabled {
		tr.Workspaces[c.Info.Desk].Tile()
	}

	// Update client desktop
	success := c.Update()
	if !success {
		return
	}

	// Add client to new workspace
	tr.Workspaces[c.Info.Desk].AddClient(c)
	if tr.Workspaces[c.Info.Desk].TilingEnabled {
		tr.Workspaces[c.Info.Desk].Tile()
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

	// Attach structure events
	xevent.ConfigureNotifyFun(func(x *xgbutil.XUtil, ev xevent.ConfigureNotifyEvent) {
		log.Debug("Client structure event [", c.Info.Class, "]")

		if tr.isTrackable(c.Win.Id) {
			tr.handleResizeClient(c)
		} else {
			tr.untrackWindow(c.Win.Id)
		}
	}).Connect(common.X, c.Win.Id)

	// Attach property events
	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		aname, _ := xprop.AtomName(common.X, ev.Atom)
		log.Debug("Client property event ", aname, " [", c.Info.Class, "]")

		if tr.isTrackable(c.Win.Id) {
			if aname == "_NET_WM_STATE" {
				tr.handleMaximizedClient(c)
				tr.handleMinimizedClient(c)
				tr.handleMoveClient(c)
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
	info := store.GetInfo(w)
	if info.Class == store.UNKNOWN {
		return false
	}

	// Check if window is allowed and inside viewport
	isAllowed := !store.IsModal(info) && !store.IsHidden(info) && !store.IsFloating(info) && !store.IsPinned(info) && !store.IsIgnored(info)
	return isAllowed && store.IsInsideViewPort(w)
}

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

var (
	timer *time.Timer // Tiling timer after window resize
)

type Tracker struct {
	Clients    map[xproto.Window]*store.Client // List of clients that are being tracked
	Workspaces map[uint]*Workspace             // List of workspaces used
}

func CreateTracker(ws map[uint]*Workspace) *Tracker {
	tr := Tracker{
		Clients:    make(map[xproto.Window]*store.Client),
		Workspaces: ws,
	}

	// Init clients
	xevent.PropertyNotifyFun(tr.handleWorkspaceUpdates).Connect(common.X, common.X.RootWin())
	tr.populateClients()

	return &tr
}

func (tr *Tracker) populateClients() {

	// Add trackable windows
	for _, w := range common.Windows {
		if tr.isTrackable(w) {
			tr.trackWindow(w)
		}
	}

	// If window is tracked, but not in client list
	for w1 := range tr.Clients {
		trackable := false
		for _, w2 := range common.Windows {
			if w1 == w2 {
				trackable = tr.isTrackable(w1)
				break
			}
		}
		if !trackable {
			tr.untrackWindow(w1)
		}
	}

	// Tile workspace
	ws := tr.Workspaces[common.CurrentDesk]
	ws.Tile()
}

func (tr *Tracker) trackWindow(w xproto.Window) {
	if tr.isTracked(w) {
		return
	}

	// Add new client
	c := store.CreateClient(w)
	tr.Clients[c.Win.Id] = c
	ws := tr.Workspaces[c.Latest.Desk]
	ws.AddClient(c)

	// Attach handlers and tile
	tr.attachHandlers(c)
	tr.tileWorkspace(c, 0)
}

func (tr *Tracker) untrackWindow(w xproto.Window) {
	if tr.isTracked(w) {
		c := tr.Clients[w]
		ws := tr.Workspaces[c.Latest.Desk]

		// Remove client
		ws.RemoveClient(c)
		xevent.Detach(common.X, w)
		delete(tr.Clients, w)
	}
}

func (tr *Tracker) tileWorkspace(c *store.Client, ms time.Duration) {
	ws := tr.Workspaces[c.Latest.Desk]

	// Tile workspace
	ws.Tile()

	// Re-tile as some applications load geometry delayed
	if ms > 0 {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(ms*time.Millisecond, ws.Tile)
	}
}

func (tr *Tracker) handleResizeClient(c *store.Client) {
	ws := tr.Workspaces[c.Latest.Desk]
	if !ws.IsEnabled() || store.IsMaximized(c.Win.Id) {
		return
	}

	// Previous dimensions
	pGeom := c.Latest.Dimensions.Geometry
	px, py, pw, ph := pGeom.Pieces()

	// Current dimensions
	cGeom, err := c.Win.DecorGeometry()
	if err != nil {
		return
	}
	cx, cy, cw, ch := cGeom.Pieces()

	// Check width/height changes and directions
	resized := math.Abs(float64(cw-pw)) > 0.0 || math.Abs(float64(ch-ph)) > 0.0
	directions := &store.Directions{Top: cy != py, Right: cx == px && cw != pw, Bottom: cy == py && ch != ph, Left: cx != px}

	// Check window lifetime
	lifetime := time.Since(c.Created)
	added := lifetime < 1000*time.Millisecond
	initialized := (math.Abs(float64(cx-px)) > 0.0 || math.Abs(float64(cy-py)) > 0.0) && added

	if resized || initialized {

		// Update proportions
		if !added {
			ws.ActiveLayout().UpdateProportions(c, directions)
		}

		// Tile workspace
		tr.tileWorkspace(c, 500)
	}
}

func (tr *Tracker) handleMoveClient(c *store.Client) {
	ws := tr.Workspaces[c.Latest.Desk]
	if !ws.IsEnabled() || store.IsMaximized(c.Win.Id) {
		return
	}

	// Previous position
	pGeom := c.Latest.Dimensions.Geometry
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
		mg := ws.ActiveLayout().GetManager()

		// Check if pointer hovers other clients
		for _, co := range mg.Clients() {
			if c.Win.Id == co.Win.Id {
				continue
			}

			// Swap moved client with hovered client
			if common.IsInsideRect(common.Pointer, co.Latest.Dimensions.Geometry) {
				mg.SwapClient(c, co)
				break
			}
		}

		// Tile workspace
		tr.tileWorkspace(c, 500)
	}
}

func (tr *Tracker) handleMaximizedClient(c *store.Client) {
	states, _ := ewmh.WmStateGet(common.X, c.Win.Id)

	// Client maximized
	for _, state := range states {
		if strings.Contains(state, "_NET_WM_STATE_MAXIMIZED") {
			ws := tr.Workspaces[c.Latest.Desk]
			if !ws.IsEnabled() {
				return
			}

			// Set fullscreen layout
			for i, l := range ws.Layouts {
				if l.GetName() == "fullscreen" {
					ws.SetLayout(uint(i))
				}
			}
			c.Activate()
			tr.tileWorkspace(c, 0)
			ShowLayout(ws)
			break
		}
	}
}

func (tr *Tracker) handleMinimizedClient(c *store.Client) {
	states, _ := ewmh.WmStateGet(common.X, c.Win.Id)

	// Client minimized
	for _, state := range states {
		if state == "_NET_WM_STATE_HIDDEN" {
			ws := tr.Workspaces[c.Latest.Desk]
			if !ws.IsEnabled() {
				return
			}

			// Untrack client
			tr.untrackWindow(c.Win.Id)
			tr.tileWorkspace(c, 0)
			break
		}
	}
}

func (tr *Tracker) handleDesktopChange(c *store.Client) {

	// Remove client from current workspace
	tr.Workspaces[c.Latest.Desk].RemoveClient(c)
	if tr.Workspaces[c.Latest.Desk].IsEnabled() {
		tr.tileWorkspace(c, 0)
	}

	// Update client desktop
	success := c.Update()
	if !success {
		return
	}

	// Add client to new workspace
	tr.Workspaces[c.Latest.Desk].AddClient(c)
	if tr.Workspaces[c.Latest.Desk].IsEnabled() {
		tr.tileWorkspace(c, 0)
	} else {
		c.Restore()
	}
}

func (tr *Tracker) handleWorkspaceUpdates(X *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
	aname, _ := xprop.AtomName(common.X, ev.Atom)
	log.Trace("Workspace update event ", aname)

	workspaceChanged := aname == "_NET_WORKAREA"
	desktopChanged := aname == "_NET_DESKTOP_LAYOUT" || aname == "_NET_DESKTOP_VIEWPORT"
	clientAdded := aname == "_NET_CLIENT_LIST" || aname == "_NET_CLIENT_LIST_STACKING"

	// Layout changed or client added
	if workspaceChanged || desktopChanged || clientAdded {
		tr.populateClients()
	}
}

func (tr *Tracker) attachHandlers(c *store.Client) {
	c.Win.Listen(xproto.EventMaskStructureNotify | xproto.EventMaskPropertyChange)

	// Attach structure events
	xevent.ConfigureNotifyFun(func(x *xgbutil.XUtil, ev xevent.ConfigureNotifyEvent) {
		log.Trace("Client structure event [", c.Latest.Class, "]")

		if tr.isTrackable(c.Win.Id) {
			tr.handleResizeClient(c)
		} else {
			tr.untrackWindow(c.Win.Id)
			tr.tileWorkspace(c, 0)
		}
	}).Connect(common.X, c.Win.Id)

	// Attach property events
	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		aname, _ := xprop.AtomName(common.X, ev.Atom)
		log.Trace("Client property event ", aname, " [", c.Latest.Class, "]")

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
			tr.tileWorkspace(c, 0)
		}
	}).Connect(common.X, c.Win.Id)
}

func (tr *Tracker) isTracked(w xproto.Window) bool {
	_, ok := tr.Clients[w]
	return ok
}

func (tr *Tracker) isTrackable(w xproto.Window) bool {
	return store.IsInsideViewPort(w) && !store.IsIgnored(w) && !store.IsSpecial(w)
}

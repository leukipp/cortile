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
	swap *Swap // Stores clients to swap after move
)

type Tracker struct {
	Clients    map[xproto.Window]*store.Client // List of clients that are being tracked
	Workspaces map[uint]*Workspace             // List of workspaces used
}

type Swap struct {
	Client1 *store.Client // Stores moving client
	Client2 *store.Client // Stores hovered client
}

func CreateTracker(ws map[uint]*Workspace) *Tracker {
	tr := Tracker{
		Clients:    make(map[xproto.Window]*store.Client),
		Workspaces: ws,
	}

	// Populate clients
	xevent.PropertyNotifyFun(tr.handleWorkspaceUpdates).Connect(common.X, common.X.RootWin())
	tr.Update()

	// Startup tiling
	if common.Config.TilingEnabled {
		for _, ws := range ws {
			ws.Tile()
		}
		ShowLayout(tr.Workspaces[common.CurrentDesk])
	}

	return &tr
}

func (tr *Tracker) Update() {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}

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
	ws.Tile()
}

func (tr *Tracker) trackWindow(w xproto.Window) {
	if tr.isTracked(w) {
		return
	}

	// Add new client
	c := store.CreateClient(w)
	tr.Clients[c.Win.Id] = c
	ws := tr.Workspaces[c.Latest.DeskNum]
	ws.AddClient(c)

	// Attach handlers and tile
	tr.attachHandlers(c)
	tr.tileWorkspace(c)
}

func (tr *Tracker) untrackWindow(w xproto.Window) {
	if tr.isTracked(w) {
		c := tr.Clients[w]
		ws := tr.Workspaces[c.Latest.DeskNum]

		// Detach events
		xevent.Detach(common.X, w)

		// Restore client
		c.Restore()

		// Remove client
		ws.RemoveClient(c)
		delete(tr.Clients, w)
	}
}

func (tr *Tracker) tileWorkspace(c *store.Client) {
	ws := tr.Workspaces[c.Latest.DeskNum]

	// Tile workspace
	ws.Tile()
}

func (tr *Tracker) handleResizeClient(c *store.Client) {
	ws := tr.Workspaces[c.Latest.DeskNum]
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

	// Check size changes
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
		tr.tileWorkspace(c)
	}
}

func (tr *Tracker) handleMoveClient(c *store.Client) {
	ws := tr.Workspaces[c.Latest.DeskNum]
	if !ws.IsEnabled() || store.IsMaximized(c.Win.Id) {
		return
	}

	// Previous position
	pGeom := c.Latest.Dimensions.Geometry
	px, py, pw, ph := pGeom.Pieces()

	// Current position
	cGeom, err := c.Win.DecorGeometry()
	if err != nil {
		return
	}
	cx, cy, cw, ch := cGeom.Pieces()

	// Check position change
	active := c.Win.Id == common.ActiveWindow
	moved := math.Abs(float64(cx-px)) > 0.0 || math.Abs(float64(cy-py)) > 0.0
	resized := math.Abs(float64(cw-pw)) > 0.0 || math.Abs(float64(ch-ph)) > 0.0

	if active && (moved && !resized) {
		mg := ws.ActiveLayout().GetManager()
		clients := mg.Clients(false)

		// Check if pointer hovers other client
		swap = nil
		for _, co := range clients {
			if c.Win.Id == co.Win.Id {
				continue
			}

			// Store moved client and hovered client
			if common.IsInsideRect(common.Pointer, co.Latest.Dimensions.Geometry) {
				swap = &Swap{
					Client1: c,
					Client2: co,
				}
				break
			}
		}
	}
}

func (tr *Tracker) handleSwapClient(c *store.Client) {
	ws := tr.Workspaces[c.Latest.DeskNum]
	if !ws.IsEnabled() || store.IsMaximized(c.Win.Id) {
		return
	}

	if swap != nil {
		mg := ws.ActiveLayout().GetManager()

		// Swap clients
		mg.SwapClient(swap.Client1, swap.Client2)
		swap = nil
	}

	// Tile workspace
	tr.tileWorkspace(c)
}

func (tr *Tracker) handleMaximizedClient(c *store.Client) {
	states, _ := ewmh.WmStateGet(common.X, c.Win.Id)

	// Client maximized
	for _, state := range states {
		if strings.Contains(state, "_NET_WM_STATE_MAXIMIZED") {
			ws := tr.Workspaces[c.Latest.DeskNum]
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
			tr.tileWorkspace(c)
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
			ws := tr.Workspaces[c.Latest.DeskNum]
			if !ws.IsEnabled() {
				return
			}

			// Untrack client
			tr.untrackWindow(c.Win.Id)
			tr.tileWorkspace(c)
			break
		}
	}
}

func (tr *Tracker) handleDesktopChange(c *store.Client) {

	// Remove client from current workspace
	tr.Workspaces[c.Latest.DeskNum].RemoveClient(c)
	if tr.Workspaces[c.Latest.DeskNum].IsEnabled() {
		tr.tileWorkspace(c)
	}

	// Update client desktop
	success := c.Update()
	if !success {
		return
	}

	// Add client to new workspace
	tr.Workspaces[c.Latest.DeskNum].AddClient(c)
	if tr.Workspaces[c.Latest.DeskNum].IsEnabled() {
		tr.tileWorkspace(c)
	} else {
		c.Restore()
	}
}

func (tr *Tracker) handleWorkspaceUpdates(X *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
	aname, _ := xprop.AtomName(common.X, ev.Atom)
	log.Trace("Workspace update event ", aname)

	clientAdded := aname == "_NET_CLIENT_LIST" || aname == "_NET_CLIENT_LIST_STACKING"
	workspaceChanged := aname == "_NET_DESKTOP_LAYOUT" || aname == "_NET_DESKTOP_VIEWPORT" || aname == "_NET_WORKAREA"

	// Client added or workspace changed
	if clientAdded || workspaceChanged {
		tr.Update()

		// Re-update as some wm minimize to outside
		time.AfterFunc(200*time.Millisecond, func() {
			if !tr.isTracked(common.ActiveWindow) {
				tr.Update()
			}
		})
	}
}

func (tr *Tracker) attachHandlers(c *store.Client) {
	c.Win.Listen(xproto.EventMaskStructureNotify | xproto.EventMaskPropertyChange | xproto.EventMaskFocusChange)

	// Attach structure events
	xevent.ConfigureNotifyFun(func(x *xgbutil.XUtil, ev xevent.ConfigureNotifyEvent) {
		log.Trace("Client structure event [", c.Latest.Class, "]")

		// Handle structure changes
		if tr.isTrackable(c.Win.Id) {
			tr.handleResizeClient(c)
			tr.handleMoveClient(c)
		} else {
			tr.Update()
		}
	}).Connect(common.X, c.Win.Id)

	// Attach property events
	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		aname, _ := xprop.AtomName(common.X, ev.Atom)
		log.Trace("Client property event ", aname, " [", c.Latest.Class, "]")

		// Handle property changes
		if tr.isTrackable(c.Win.Id) {
			if aname == "_NET_WM_STATE" {
				tr.handleMaximizedClient(c)
				tr.handleMinimizedClient(c)
			} else if aname == "_NET_WM_DESKTOP" {
				tr.handleDesktopChange(c)
			}
		} else {
			tr.Update()
		}
	}).Connect(common.X, c.Win.Id)

	// Attach focus events
	xevent.FocusInFun(func(x *xgbutil.XUtil, ev xevent.FocusInEvent) {
		log.Trace("Client focus event [", c.Latest.Class, "]")

		// Wait for structure changes
		time.AfterFunc(200*time.Millisecond, func() {
			if ev.Mode == xproto.NotifyModeUngrab {
				tr.handleSwapClient(c)
			}
		})
	}).Connect(common.X, c.Win.Id)
}

func (tr *Tracker) isTracked(w xproto.Window) bool {
	_, ok := tr.Clients[w]
	return ok
}

func (tr *Tracker) isTrackable(w xproto.Window) bool {
	return store.IsInsideViewPort(w) && !store.IsIgnored(w) && !store.IsSpecial(w)
}

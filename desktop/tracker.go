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
	Workspaces map[Location]*Workspace         // List of workspaces per location
}

type Location struct {
	DeskNum   uint // Workspace desktop number
	ScreenNum uint // Workspace screen number
}

type Swap struct {
	Client1 *store.Client // Stores moving client
	Client2 *store.Client // Stores hovered client
}

func CreateTracker(ws map[Location]*Workspace) *Tracker {
	tr := Tracker{
		Clients:    make(map[xproto.Window]*store.Client),
		Workspaces: ws,
	}

	// Attach to state update events
	common.OnStateUpdate(tr.onStateUpdate)

	// Populate clients
	tr.Update()

	// Startup tiling
	if common.Config.TilingEnabled {
		ShowLayout(tr.ActiveWorkspace())
	}

	return &tr
}

func (tr *Tracker) Update() {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return
	}

	// Map trackable windows
	trackable := make(map[xproto.Window]bool)
	for _, w := range common.Windows {
		trackable[w] = tr.isTrackable(w)
	}

	// Remove untrackable windows
	for w := range tr.Clients {
		if !trackable[w] {
			tr.untrackWindow(w)
		}
	}

	// Add trackable windows
	for _, w := range common.Windows {
		if trackable[w] {
			tr.trackWindow(w)
		}
	}

	// Tile workspace
	ws.Tile()
}

func (tr *Tracker) Reset() {

	// Reset client list
	for w := range tr.Clients {
		tr.untrackWindow(w)
	}

	// Reset workspaces
	tr.Workspaces = CreateWorkspaces()
}

func (tr *Tracker) ActiveWorkspace() *Workspace {
	return tr.Workspaces[Location{DeskNum: common.CurrentDesk, ScreenNum: common.CurrentScreen}]
}

func (tr *Tracker) ClientWorkspace(c *store.Client) *Workspace {
	return tr.Workspaces[Location{DeskNum: c.Latest.DeskNum, ScreenNum: c.Latest.ScreenNum}]
}

func (tr *Tracker) trackWindow(w xproto.Window) {
	if tr.isTracked(w) {
		return
	}

	// Add new client
	c := store.CreateClient(w)
	tr.Clients[c.Win.Id] = c
	ws := tr.ClientWorkspace(c)
	ws.AddClient(c)

	// Attach handlers and tile
	tr.attachHandlers(c)
	tr.tileWorkspace(c)
}

func (tr *Tracker) untrackWindow(w xproto.Window) {
	if !tr.isTracked(w) {
		return
	}

	// Client and workspace
	c := tr.Clients[w]
	ws := tr.ClientWorkspace(c)

	// Detach events
	xevent.Detach(common.X, w)

	// Restore client
	c.Restore()

	// Remove client
	ws.RemoveClient(c)
	delete(tr.Clients, w)
}

func (tr *Tracker) tileWorkspace(c *store.Client) {
	ws := tr.ClientWorkspace(c)

	// Tile workspace
	ws.Tile()
}

func (tr *Tracker) handleResizeClient(c *store.Client) {
	ws := tr.ClientWorkspace(c)
	if !tr.isTracked(c.Win.Id) || !ws.IsEnabled() || store.IsMaximized(c.Win.Id) {
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
	moved := math.Abs(float64(cx-px)) > 0.0 || math.Abs(float64(cy-py)) > 0.0
	resized := math.Abs(float64(cw-pw)) > 0.0 || math.Abs(float64(ch-ph)) > 0.0
	directions := &store.Directions{Top: cy != py, Right: cx == px && cw != pw, Bottom: cy == py && ch != ph, Left: cx != px}

	// Check window lifetime
	lifetime := time.Since(c.Created)
	added := lifetime < 1000*time.Millisecond
	initialized := moved && added

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
	ws := tr.ClientWorkspace(c)
	if !tr.isTracked(c.Win.Id) || !ws.IsEnabled() || store.IsMaximized(c.Win.Id) {
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

	if active && moved && !resized {
		mg := ws.ActiveLayout().GetManager()
		swap = nil

		// Check if pointer hovers other client
		clients := mg.Clients(false)
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
				return
			}
		}
	}
}

func (tr *Tracker) handleSwapClient(c *store.Client) {
	ws := tr.ClientWorkspace(c)
	if !tr.isTracked(c.Win.Id) || !ws.IsEnabled() || store.IsMaximized(c.Win.Id) {
		return
	}

	if swap != nil {
		mg := ws.ActiveLayout().GetManager()

		// Swap clients on same desktop and screen
		mg.SwapClient(swap.Client1, swap.Client2)

		// Reset swap
		swap = nil
	}

	// Tile workspace
	tr.tileWorkspace(c)
}

func (tr *Tracker) handleMaximizedClient(c *store.Client) {
	if !tr.isTracked(c.Win.Id) {
		return
	}

	// Client maximized
	states, _ := ewmh.WmStateGet(common.X, c.Win.Id)
	for _, state := range states {
		if strings.Contains(state, "_NET_WM_STATE_MAXIMIZED") {
			ws := tr.ClientWorkspace(c)
			if !ws.IsEnabled() {
				return
			}

			// Set fullscreen layout
			for i, l := range ws.Layouts {
				if l.GetName() == "fullscreen" {
					ws.SetLayout(uint(i))
				}
			}
			tr.tileWorkspace(c)
			c.Activate()

			ShowLayout(ws)
			break
		}
	}
}

func (tr *Tracker) handleMinimizedClient(c *store.Client) {
	if !tr.isTracked(c.Win.Id) {
		return
	}

	// Client minimized
	states, _ := ewmh.WmStateGet(common.X, c.Win.Id)
	for _, state := range states {
		if state == "_NET_WM_STATE_HIDDEN" {
			ws := tr.ClientWorkspace(c)
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

func (tr *Tracker) handleWorkspaceChange(c *store.Client) {

	// Remove client from current workspace
	ws := tr.ClientWorkspace(c)
	ws.RemoveClient(c)
	if ws.IsEnabled() {
		tr.tileWorkspace(c)
	}

	// Update client desktop and screen
	if !tr.isTrackable(c.Win.Id) {
		return
	}
	c.Update()

	// Add client to new workspace
	ws = tr.ClientWorkspace(c)
	ws.AddClient(c)
	if ws.IsEnabled() {
		tr.tileWorkspace(c)
	} else {
		c.Restore()
	}
}

func (tr *Tracker) handleDesktopChange(c *store.Client) {
	if !tr.isTracked(c.Win.Id) {
		return
	}
	tr.handleWorkspaceChange(c)
}

func (tr *Tracker) handleScreenChange(c *store.Client) {
	if !tr.isTracked(c.Win.Id) || c.Latest.ScreenNum == common.CurrentScreen {
		return
	}
	tr.handleWorkspaceChange(c)
}

func (tr *Tracker) onStateUpdate(aname string) {
	clientAdded := common.IsInList(aname, []string{"_NET_CLIENT_LIST_STACKING"})
	workspacesChanged := common.DeskCount*common.ScreenCount != uint(len(tr.Workspaces))
	viewportChanged := common.IsInList(aname, []string{"_NET_NUMBER_OF_DESKTOPS", "_NET_DESKTOP_LAYOUT", "_NET_DESKTOP_GEOMETRY", "_NET_DESKTOP_VIEWPORT", "_NET_WORKAREA"})

	// Number of desktops or screens changed
	if viewportChanged && workspacesChanged {
		tr.Reset()
	}

	// Viewport changed or client added
	if viewportChanged || clientAdded {
		tr.Update()
	}
}

func (tr *Tracker) attachHandlers(c *store.Client) {
	c.Win.Listen(xproto.EventMaskStructureNotify | xproto.EventMaskPropertyChange | xproto.EventMaskFocusChange)

	// Attach structure events
	xevent.ConfigureNotifyFun(func(x *xgbutil.XUtil, ev xevent.ConfigureNotifyEvent) {
		log.Trace("Client structure event [", c.Latest.Class, "]")

		// Handle structure events
		tr.handleResizeClient(c)
		tr.handleMoveClient(c)
	}).Connect(common.X, c.Win.Id)

	// Attach property events
	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		aname, _ := xprop.AtomName(common.X, ev.Atom)
		log.Trace("Client property event ", aname, " [", c.Latest.Class, "]")

		// Handle property events
		if aname == "_NET_WM_STATE" {
			tr.handleMaximizedClient(c)
			tr.handleMinimizedClient(c)
		} else if aname == "_NET_WM_DESKTOP" {
			tr.handleDesktopChange(c)
		}
	}).Connect(common.X, c.Win.Id)

	// Attach focus events
	xevent.FocusInFun(func(x *xgbutil.XUtil, ev xevent.FocusInEvent) {
		log.Trace("Client focus event [", c.Latest.Class, "]")

		// Handle ungrab events
		if ev.Mode == xproto.NotifyModeUngrab {
			tr.handleScreenChange(c)

			// Wait for structure events
			time.AfterFunc(100*time.Millisecond, func() {
				tr.handleSwapClient(c)
			})
		}
	}).Connect(common.X, c.Win.Id)
}

func (tr *Tracker) isTracked(w xproto.Window) bool {
	_, ok := tr.Clients[w]
	return ok
}

func (tr *Tracker) isTrackable(w xproto.Window) bool {
	info := store.GetInfo(w)
	return !store.IsSpecial(info) && !store.IsIgnored(info)
}

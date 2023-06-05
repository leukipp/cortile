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
	Workspaces map[Location]*Workspace         // List of workspaces per location
	Handler    *Handler                        // Helper for event handlers
}

type Location struct {
	DeskNum   uint // Workspace desktop number
	ScreenNum uint // Workspace screen number
}

type Handler struct {
	Resize *ResizeHandler // Stores variables of resize handler
	Move   *MoveHandler   // Stores variables of move handler
}

type ResizeHandler struct {
	Fired  bool          // Indicates fired resize event
	Client *ResizeClient // Stores client for proportion change
}

type ResizeClient struct {
	Active bool          // Indicates active client resize
	Source *store.Client // Stores user resized client
}

type MoveHandler struct {
	Fired  bool        // Indicates fired move event
	Client *SwapClient // Stores clients for window swap
	Screen *SwapScreen // Stores client for screen change
}

type SwapClient struct {
	Active bool          // Indicates active client swap
	Source *store.Client // Stores moving client for window swap
	Target *store.Client // Stores hovered client for window swap
}

type SwapScreen struct {
	Active bool          // Indicates active screen change
	Source *store.Client // Stores moving client for screen change
}

func CreateTracker(ws map[Location]*Workspace) *Tracker {
	tr := Tracker{
		Clients:    make(map[xproto.Window]*store.Client),
		Workspaces: ws,
		Handler: &Handler{
			Resize: &ResizeHandler{
				Client: &ResizeClient{},
			},
			Move: &MoveHandler{
				Client: &SwapClient{},
				Screen: &SwapScreen{},
			},
		},
	}

	// Attach to root events
	common.OnStateUpdate(tr.onStateUpdate)
	common.OnPointerUpdate(tr.onPointerUpdate)

	// Populate clients on startup
	if common.Config.TilingEnabled {
		tr.Update(true)
		ShowLayout(tr.ActiveWorkspace())
	}

	return &tr
}

func (tr *Tracker) Update(tile bool) {
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
			if tr.untrackWindow(w) {
				tile = true
			}
		}
	}

	// Add trackable windows
	for _, w := range common.Windows {
		if trackable[w] {
			if tr.trackWindow(w) {
				tile = true
			}
		}
	}

	// Tile workspace
	if tile {
		ws.Tile()
	}
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

func (tr *Tracker) trackWindow(w xproto.Window) bool {
	if tr.isTracked(w) {
		return false
	}

	// Add new client
	c := store.CreateClient(w)
	tr.Clients[c.Win.Id] = c
	ws := tr.ClientWorkspace(c)
	ws.AddClient(c)

	// Attach handlers
	tr.attachHandlers(c)
	ws.Tile()

	return true
}

func (tr *Tracker) untrackWindow(w xproto.Window) bool {
	if !tr.isTracked(w) {
		return false
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

	return true
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

			// Tile workspace
			ws.Tile()
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
			ws.Tile()
			break
		}
	}
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
		al := ws.ActiveLayout()

		// Set client resize event
		if !tr.Handler.Resize.Fired {
			tr.Handler.Resize.Client = &ResizeClient{Active: true, Source: c}
		}
		tr.Handler.Resize.Fired = true

		if !added {

			// Set client resize lock
			if tr.Handler.Resize.Client.Active {
				tr.Handler.Resize.Client.Source.Lock()
			}

			// Update proportions
			al.UpdateProportions(c, directions)
		}

		// Tile workspace
		ws.Tile()
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

		// Set client move event
		tr.Handler.Move.Fired = true

		// Check if pointer hovers another client
		tr.Handler.Move.Client.Active = false
		for _, co := range mg.Clients(false) {
			if c == nil || co == nil || c.Win.Id == co.Win.Id {
				continue
			}

			// Store moved client and hovered client
			if common.IsInsideRect(common.CurrentPointer, co.Latest.Dimensions.Geometry) {
				tr.Handler.Move.Client = &SwapClient{Active: true, Source: c, Target: co}
				break
			}
		}

		// Check if pointer moves to another screen
		tr.Handler.Move.Screen.Active = false
		if c.Latest.ScreenNum != common.CurrentScreen {
			tr.Handler.Move.Screen = &SwapScreen{Active: true, Source: c}
		}
	}
}

func (tr *Tracker) handleSwapClient(c *store.Client) {
	ws := tr.ClientWorkspace(c)
	if !tr.isTracked(c.Win.Id) || !ws.IsEnabled() || store.IsMaximized(c.Win.Id) {
		return
	}

	// Swap clients on same desktop and screen
	mg := ws.ActiveLayout().GetManager()
	mg.SwapClient(tr.Handler.Move.Client.Source, tr.Handler.Move.Client.Target)

	// Reset client swapping event
	tr.Handler.Move.Client.Active = false

	// Tile workspace
	ws.Tile()
}

func (tr *Tracker) handleWorkspaceChange(c *store.Client) {
	if !tr.isTracked(c.Win.Id) {
		return
	}

	// Remove client from current workspace
	ws := tr.ClientWorkspace(c)
	ws.RemoveClient(c)
	if ws.IsEnabled() {
		ws.Tile()
	}

	// Reset screen swapping event
	tr.Handler.Move.Screen.Active = false

	// Update client desktop and screen
	if !tr.isTrackable(c.Win.Id) {
		return
	}
	c.Update()

	// Add client to new workspace
	ws = tr.ClientWorkspace(c)
	ws.AddClient(c)
	if ws.IsEnabled() {
		ws.Tile()
	} else {
		c.Restore()
	}
}

func (tr *Tracker) onStateUpdate(aname string) {
	workspacesChanged := common.DeskCount*common.ScreenCount != uint(len(tr.Workspaces))
	viewportChanged := common.IsInList(aname, []string{"_NET_NUMBER_OF_DESKTOPS", "_NET_DESKTOP_LAYOUT", "_NET_DESKTOP_GEOMETRY", "_NET_DESKTOP_VIEWPORT", "_NET_WORKAREA"})
	clientsChanged := common.IsInList(aname, []string{"_NET_CLIENT_LIST_STACKING", "_NET_ACTIVE_WINDOW"})

	// Number of desktops or screens changed
	if workspacesChanged {
		tr.Reset()
	}

	// Viewport changed or clients changed
	if viewportChanged || clientsChanged {
		tr.Update(viewportChanged)
	}
}

func (tr *Tracker) onPointerUpdate(button uint16) {

	// Window resized
	if tr.Handler.Resize.Client.Active {
		tr.Handler.Resize.Client.Active = false
	}

	// Window moved over another window
	if tr.Handler.Move.Client.Active {
		tr.handleSwapClient(tr.Handler.Move.Client.Source)
	}

	// Window moved to another screen
	if tr.Handler.Move.Screen.Active {
		tr.handleWorkspaceChange(tr.Handler.Move.Screen.Source)
	}

	// Reset client resize and move events
	if tr.Handler.Resize.Fired || tr.Handler.Move.Fired {
		tr.Handler.Resize.Fired = false
		tr.Handler.Move.Fired = false

		// Tile workspace
		tr.ActiveWorkspace().Tile()
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
			tr.handleWorkspaceChange(c)
		}
	}).Connect(common.X, c.Win.Id)

	// Attach focus in events
	xevent.FocusInFun(func(x *xgbutil.XUtil, ev xevent.FocusInEvent) {
		log.Trace("Client focus in event [", c.Latest.Class, "]")

		// Update active window
		common.ActiveWindow, _ = ewmh.ActiveWindowGet(common.X)
	}).Connect(common.X, c.Win.Id)

	// Attach focus out events
	xevent.FocusOutFun(func(x *xgbutil.XUtil, ev xevent.FocusOutEvent) {
		log.Trace("Client focus out event [", c.Latest.Class, "]")

		// Update active window
		common.ActiveWindow, _ = ewmh.ActiveWindowGet(common.X)
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

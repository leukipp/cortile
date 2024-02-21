package desktop

import (
	"strings"
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

type Tracker struct {
	Clients    map[xproto.Window]*store.Client // List of clients that are being tracked
	Workspaces map[store.Location]*Workspace   // List of workspaces per location
	Action     chan string                     // Event channel for actions
	Handler    *Handler                        // Helper for event handlers
}

type Handler struct {
	Timer        *time.Timer    // Timer to handle delayed structure events
	ResizeClient *HandlerClient // Stores client for proportion change
	MoveClient   *HandlerClient // Stores client for tiling after move
	SwapClient   *HandlerClient // Stores clients for window swap
	SwapScreen   *HandlerClient // Stores client for screen swap
}

type HandlerClient struct {
	Active bool          // Indicates active handler event
	Source *store.Client // Stores moving/resizing client
	Target *store.Client // Stores hovered client
}

func CreateTracker(ws map[store.Location]*Workspace) *Tracker {
	tr := Tracker{
		Clients:    make(map[xproto.Window]*store.Client),
		Workspaces: ws,
		Action:     make(chan string),
		Handler: &Handler{
			ResizeClient: &HandlerClient{},
			MoveClient:   &HandlerClient{},
			SwapClient:   &HandlerClient{},
			SwapScreen:   &HandlerClient{},
		},
	}

	// Attach to root events
	store.OnStateUpdate(tr.onStateUpdate)
	store.OnPointerUpdate(tr.onPointerUpdate)

	// Populate clients on startup
	if common.Config.TilingEnabled {
		tr.Update()
	}

	return &tr
}

func (tr *Tracker) Update() {
	ws := tr.ActiveWorkspace()
	if ws.Disabled() {
		return
	}
	log.Debug("Update trackable clients [", len(tr.Clients), "/", len(store.Windows), "]")

	// Map trackable windows
	trackable := make(map[xproto.Window]bool)
	for _, w := range store.Windows {
		trackable[w] = tr.isTrackable(w)
	}

	// Remove untrackable windows
	for w := range tr.Clients {
		if !trackable[w] {
			tr.untrackWindow(w)
		}
	}

	// Add trackable windows
	for _, w := range store.Windows {
		if trackable[w] {
			tr.trackWindow(w)
		}
	}
}

func (tr *Tracker) Reset() {
	log.Debug("Reset trackable clients [", len(tr.Clients), "/", len(store.Windows), "]")

	// Reset client list
	for w := range tr.Clients {
		tr.untrackWindow(w)
	}

	// Reset workspaces
	tr.Workspaces = CreateWorkspaces()
}

func (tr *Tracker) ActiveWorkspace() *Workspace {
	location := store.Location{DeskNum: store.CurrentDesk, ScreenNum: store.CurrentScreen}

	// Validate active workspace
	ws := tr.Workspaces[location]
	if ws == nil {
		log.Warn("Invalid active workspace [workspace-", location.DeskNum, "-", location.ScreenNum, "]")
	}

	return ws
}

func (tr *Tracker) ClientWorkspace(c *store.Client) *Workspace {
	location := store.Location{DeskNum: c.Latest.Location.DeskNum, ScreenNum: c.Latest.Location.ScreenNum}

	// Validate client workspace
	ws := tr.Workspaces[location]
	if ws == nil {
		log.Warn("Invalid client workspace [workspace-", location.DeskNum, "-", location.ScreenNum, "]")
	}

	return ws
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
	xevent.Detach(store.X, w)

	// Restore client
	c.Restore(store.Latest)

	// Remove client
	ws.RemoveClient(c)
	delete(tr.Clients, w)
	ws.Tile()

	return true
}

func (tr *Tracker) handleMaximizedClient(c *store.Client) {
	if !tr.isTracked(c.Win.Id) {
		return
	}

	// Client maximized
	states, _ := ewmh.WmStateGet(store.X, c.Win.Id)
	for _, state := range states {
		if strings.HasPrefix(state, "_NET_WM_STATE_MAXIMIZED") {
			ws := tr.ClientWorkspace(c)
			if ws.Disabled() {
				return
			}
			log.Debug("Client maximized handler fired [", c.Latest.Class, "]")

			// Update client states
			c.Update()

			// Set fullscreen layout
			c.UnMaximize()
			tr.Action <- "layout_fullscreen"
			c.Activate()

			break
		}
	}
}

func (tr *Tracker) handleMinimizedClient(c *store.Client) {
	if !tr.isTracked(c.Win.Id) {
		return
	}

	// Client minimized
	states, _ := ewmh.WmStateGet(store.X, c.Win.Id)
	for _, state := range states {
		if state == "_NET_WM_STATE_HIDDEN" {
			ws := tr.ClientWorkspace(c)
			if ws.Disabled() {
				return
			}
			log.Debug("Client minimized handler fired [", c.Latest.Class, "]")

			// Untrack client
			tr.untrackWindow(c.Win.Id)

			break
		}
	}
}

func (tr *Tracker) handleResizeClient(c *store.Client) {
	ws := tr.ClientWorkspace(c)
	if ws.Disabled() || !tr.isTracked(c.Win.Id) || store.IsMaximized(c.Win.Id) {
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
	resized := cw != pw || ch != ph
	moved := (cx != px || cy != py) && (cw == pw && ch == ph)
	added := time.Since(c.Created) < 1000*time.Millisecond

	if resized && !moved && !tr.Handler.MoveClient.Active {
		al := ws.ActiveLayout()

		// Set client resize event
		if !tr.Handler.ResizeClient.Active {
			tr.Handler.ResizeClient = &HandlerClient{Active: true, Source: c}
		}
		log.Debug("Client resize handler fired [", c.Latest.Class, "]")

		// Check window lifetime
		if !added {

			// Set client resize lock
			if tr.Handler.ResizeClient.Active {
				tr.Handler.ResizeClient.Source.Lock()
				log.Debug("Client resize handler active [", c.Latest.Class, "]")
			}

			// Update proportions
			dir := &store.Directions{
				Top:    cy != py,
				Right:  cx == px && cw != pw,
				Bottom: cy == py && ch != ph,
				Left:   cx != px,
			}
			al.UpdateProportions(c, dir)
		}

		// Tile workspace
		ws.Tile()
	}
}

func (tr *Tracker) handleMoveClient(c *store.Client) {
	ws := tr.ClientWorkspace(c)
	if !tr.isTracked(c.Win.Id) || store.IsMaximized(c.Win.Id) {
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

	// Check position changes
	moved := cx != px || cy != py
	resized := cw != pw || ch != ph
	active := c.Win.Id == store.ActiveWindow

	if active && moved && !resized && !tr.Handler.ResizeClient.Active {
		mg := ws.ActiveLayout().GetManager()
		pt := store.PointerUpdate(store.X)

		// Set client move event
		if !tr.Handler.MoveClient.Active {
			tr.Handler.MoveClient = &HandlerClient{Active: true, Source: c}
		}
		log.Debug("Client move handler fired [", c.Latest.Class, "]")

		// Check if pointer hovers another client
		tr.Handler.SwapClient.Active = false
		for _, co := range mg.Clients(store.Visible) {
			if co == nil || c.Win.Id == co.Win.Id {
				continue
			}

			// Store moved client and hovered client
			if common.IsInsideRect(pt, co.Latest.Dimensions.Geometry.Rect) {
				tr.Handler.SwapClient = &HandlerClient{Active: true, Source: c, Target: co}
				log.Debug("Client move handler active [", c.Latest.Class, "-", co.Latest.Class, "]")
				break
			}
		}

		// Check if pointer moves to another screen
		tr.Handler.SwapScreen.Active = false
		if c.Latest.Location.ScreenNum != store.CurrentScreen {
			tr.Handler.SwapScreen = &HandlerClient{Active: true, Source: c}
		}
	}
}

func (tr *Tracker) handleSwapClient(c *store.Client) {
	ws := tr.ClientWorkspace(c)
	if !tr.isTracked(c.Win.Id) {
		return
	}
	log.Debug("Client swap handler fired [", tr.Handler.SwapClient.Source.Latest.Class, "-", tr.Handler.SwapClient.Target.Latest.Class, "]")

	// Swap clients on same desktop and screen
	mg := ws.ActiveLayout().GetManager()
	mg.SwapClient(tr.Handler.SwapClient.Source, tr.Handler.SwapClient.Target)

	// Reset client swapping event
	tr.Handler.SwapClient.Active = false

	// Tile workspace
	ws.Tile()
}

func (tr *Tracker) handleWorkspaceChange(c *store.Client) {
	if !tr.isTracked(c.Win.Id) {
		return
	}
	log.Debug("Client workspace handler fired [", c.Latest.Class, "]")

	// Remove client from current workspace
	ws := tr.ClientWorkspace(c)
	mg := ws.ActiveLayout().GetManager()
	master := mg.IsMaster(c)
	ws.RemoveClient(c)

	// Tile current workspace
	if ws.Enabled() {
		ws.Tile()
	}

	// Update client desktop and screen
	if !tr.isTrackable(c.Win.Id) {
		return
	}
	c.Update()

	// Add client to new workspace
	ws = tr.ClientWorkspace(c)
	if tr.Handler.SwapScreen.Active && tr.ActiveWorkspace().Enabled() {
		ws = tr.ActiveWorkspace()
	}
	mg = ws.ActiveLayout().GetManager()
	ws.AddClient(c)
	if master {
		mg.MakeMaster(c)
	}

	// Tile new workspace
	if ws.Enabled() {
		ws.Tile()
	} else {
		c.Restore(store.Latest)
	}

	// Reset screen swapping event
	tr.Handler.SwapScreen.Active = false
}

func (tr *Tracker) unlockClients() {
	ws := tr.ActiveWorkspace()
	mg := ws.ActiveLayout().GetManager()

	// Unlock clients
	for _, c := range mg.Clients(store.Stacked) {
		if c == nil {
			continue
		}
		c.UnLock()
	}
}

func (tr *Tracker) onStateUpdate(aname string) {
	viewportChanged := common.IsInList(aname, []string{"_NET_NUMBER_OF_DESKTOPS", "_NET_DESKTOP_LAYOUT", "_NET_DESKTOP_GEOMETRY", "_NET_DESKTOP_VIEWPORT", "_NET_WORKAREA"})
	clientsChanged := common.IsInList(aname, []string{"_NET_CLIENT_LIST_STACKING", "_NET_ACTIVE_WINDOW"})

	workspacesChanged := store.DeskCount*store.ScreenCount != uint(len(tr.Workspaces))
	workspaceChanged := common.IsInList(aname, []string{"_NET_CURRENT_DESKTOP"})

	// Number of desktops or screens changed
	if workspacesChanged {
		tr.Reset()
	}

	// Active desktop changed
	if workspaceChanged {
		for _, c := range tr.Clients {
			sticky := common.IsInList("_NET_WM_STATE_STICKY", c.Latest.States)
			if sticky && c.Latest.Location.DeskNum != store.CurrentDesk {
				ewmh.WmDesktopSet(store.X, c.Win.Id, ^uint(0))
			}
		}
	}

	// Viewport changed or clients changed
	if viewportChanged || clientsChanged {

		// Deactivate handlers
		tr.Handler.ResizeClient.Active = false
		tr.Handler.MoveClient.Active = false
		tr.Handler.SwapClient.Active = false
		tr.Handler.SwapScreen.Active = false

		// Unlock clients
		tr.unlockClients()

		// Update trackable clients
		tr.Update()
	}
}

func (tr *Tracker) onPointerUpdate(button uint16) {
	release := button == 0

	// Reset timer
	if tr.Handler.Timer != nil {
		tr.Handler.Timer.Stop()
	}

	// Wait on button release
	var t time.Duration = 0
	if release {
		t = 50
	}

	// Wait for structure events
	tr.Handler.Timer = time.AfterFunc(t*time.Millisecond, func() {

		// Window moved to another screen
		if tr.Handler.SwapScreen.Active {
			tr.handleWorkspaceChange(tr.Handler.SwapScreen.Source)
		}

		// Window moved over another window
		if tr.Handler.SwapClient.Active {
			tr.handleSwapClient(tr.Handler.SwapClient.Source)
		}

		// Window moved or resized
		if tr.Handler.MoveClient.Active || tr.Handler.ResizeClient.Active {
			tr.Handler.MoveClient.Active = false
			tr.Handler.ResizeClient.Active = false

			// Unlock clients
			tr.unlockClients()

			// Tile workspace
			if release {
				tr.ActiveWorkspace().Tile()
			}
		}
	})
}

func (tr *Tracker) attachHandlers(c *store.Client) {
	c.Win.Listen(xproto.EventMaskStructureNotify | xproto.EventMaskPropertyChange | xproto.EventMaskFocusChange)

	// Attach structure events
	xevent.ConfigureNotifyFun(func(x *xgbutil.XUtil, ev xevent.ConfigureNotifyEvent) {
		log.Trace("Client structure event [", c.Latest.Class, "]")

		// Handle structure events
		tr.handleResizeClient(c)
		tr.handleMoveClient(c)
	}).Connect(store.X, c.Win.Id)

	// Attach property events
	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		aname, _ := xprop.AtomName(store.X, ev.Atom)
		log.Trace("Client property event ", aname, " [", c.Latest.Class, "]")

		// Handle property events
		if aname == "_NET_WM_STATE" {
			tr.handleMaximizedClient(c)
			tr.handleMinimizedClient(c)
		} else if aname == "_NET_WM_DESKTOP" {
			tr.handleWorkspaceChange(c)
		}
	}).Connect(store.X, c.Win.Id)

	// Attach focus in events
	xevent.FocusInFun(func(x *xgbutil.XUtil, ev xevent.FocusInEvent) {
		log.Trace("Client focus in event [", c.Latest.Class, "]")

		// Update active window
		store.ActiveWindow = store.ActiveWindowGet(store.X)
	}).Connect(store.X, c.Win.Id)

	// Attach focus out events
	xevent.FocusOutFun(func(x *xgbutil.XUtil, ev xevent.FocusOutEvent) {
		log.Trace("Client focus out event [", c.Latest.Class, "]")

		// Update active window
		store.ActiveWindow = store.ActiveWindowGet(store.X)
	}).Connect(store.X, c.Win.Id)
}

func (tr *Tracker) isTracked(w xproto.Window) bool {
	_, ok := tr.Clients[w]
	return ok
}

func (tr *Tracker) isTrackable(w xproto.Window) bool {
	info := store.GetInfo(w)
	return !store.IsSpecial(info) && !store.IsIgnored(info)
}

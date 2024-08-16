package desktop

import (
	"time"

	"github.com/jezek/xgb/xproto"

	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/xevent"
	"github.com/jezek/xgbutil/xprop"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

type Tracker struct {
	Clients    map[xproto.Window]*store.Client // List of tracked clients
	Workspaces map[store.Location]*Workspace   // List of workspaces per location
	Channels   *Channels                       // Helper for channel communication
	Handlers   *Handlers                       // Helper for event handlers

}
type Channels struct {
	Event  chan string // Channel for events
	Action chan string // Channel for actions
}

type Handlers struct {
	Timer        *time.Timer // Timer to handle delayed structure events
	ResizeClient *Handler    // Stores client for proportion change
	MoveClient   *Handler    // Stores client for tiling after move
	SwapClient   *Handler    // Stores clients for window swap
	SwapScreen   *Handler    // Stores client for screen swap
}

type Handler struct {
	Dragging bool        // Indicates pointer dragging event
	Source   interface{} // Stores moved/resized client
	Target   interface{} // Stores client/workspace
}

func (h *Handler) Active() bool {
	return h.Source != nil
}

func (h *Handler) Reset() {
	*h = Handler{}
}

func CreateTracker() *Tracker {
	tr := Tracker{
		Clients:    make(map[xproto.Window]*store.Client),
		Workspaces: CreateWorkspaces(),
		Channels: &Channels{
			Event:  make(chan string),
			Action: make(chan string),
		},
		Handlers: &Handlers{
			ResizeClient: &Handler{},
			MoveClient:   &Handler{},
			SwapClient:   &Handler{},
			SwapScreen:   &Handler{},
		},
	}

	// Attach to root events
	store.OnStateUpdate(tr.onStateUpdate)
	store.OnPointerUpdate(tr.onPointerUpdate)

	return &tr
}

func (tr *Tracker) Update() {
	ws := tr.ActiveWorkspace()
	if ws.TilingDisabled() {
		return
	}
	log.Debug("Update trackable clients [", len(tr.Clients), "/", len(store.Windows.Stacked), "]")

	// Map trackable windows
	trackable := make(map[xproto.Window]bool)
	for _, w := range store.Windows.Stacked {
		trackable[w.Id] = tr.isTrackable(w.Id)
	}

	// Remove untrackable windows
	for w := range tr.Clients {
		if !trackable[w] {
			tr.untrackWindow(w)
		}
	}

	// Add trackable windows
	for _, w := range store.Windows.Stacked {
		if trackable[w.Id] {
			tr.trackWindow(w.Id)
		}
	}
}

func (tr *Tracker) Reset() {
	log.Debug("Reset trackable clients [", len(tr.Clients), "/", len(store.Windows.Stacked), "]")

	// Reset client list
	for w := range tr.Clients {
		tr.untrackWindow(w)
	}

	// Reset workspaces
	tr.Workspaces = CreateWorkspaces()

	// Communicate workplace change
	tr.Channels.Event <- "workplace_change"
}

func (tr *Tracker) Write() {

	// Write client cache
	for _, c := range tr.Clients {
		c.Write()
	}

	// Write workspace cache
	for _, ws := range tr.Workspaces {
		ws.Write()
	}

	// Communicate windows change
	tr.Channels.Event <- "windows_change"
}

func (tr *Tracker) Tile(ws *Workspace) {
	if ws.TilingDisabled() {
		return
	}

	// Tile workspace
	ws.Tile()

	// Communicate clients change
	tr.Channels.Event <- "clients_change"

	// Communicate workspaces change
	tr.Channels.Event <- "workspaces_change"
}

func (tr *Tracker) Restore(ws *Workspace, flag uint8) {

	// Restore workspace
	ws.Restore(flag)

	// Communicate clients change
	tr.Channels.Event <- "clients_change"

	// Communicate workspaces change
	tr.Channels.Event <- "workspaces_change"
}

func (tr *Tracker) ActiveWorkspace() *Workspace {
	return tr.WorkspaceAt(store.Workplace.CurrentDesktop, store.Workplace.CurrentScreen)
}

func (tr *Tracker) ClientWorkspace(c *store.Client) *Workspace {
	return tr.WorkspaceAt(c.Latest.Location.Desktop, c.Latest.Location.Screen)
}

func (tr *Tracker) WorkspaceAt(desktop uint, screen uint) *Workspace {
	location := store.Location{Desktop: desktop, Screen: screen}

	// Validate workspace
	ws := tr.Workspaces[location]
	if ws == nil {
		log.Warn("Invalid workspace [workspace-", location.Desktop, "-", location.Screen, "]")
	}

	return ws
}

func (tr *Tracker) ClientAt(ws *Workspace, p common.Point) *store.Client {
	mg := ws.ActiveLayout().GetManager()

	// Check if point hovers visible client
	for _, c := range mg.Clients(store.Visible) {
		if c == nil {
			continue
		}
		if common.IsInsideRect(p, c.Latest.Dimensions.Geometry) {
			return c
		}
	}

	return nil
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

func (tr *Tracker) trackWindow(w xproto.Window) bool {
	if tr.isTracked(w) {
		return false
	}

	// Add new client
	c := store.CreateClient(w)
	tr.Clients[c.Window.Id] = c
	ws := tr.ClientWorkspace(c)
	ws.AddClient(c)

	// Attach handlers
	tr.attachHandlers(c)
	tr.Tile(ws)

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

	// Tile workspace
	tr.Tile(ws)

	return true
}

func (tr *Tracker) handleMaximizedClient(c *store.Client) {
	if !tr.isTracked(c.Window.Id) {
		return
	}

	// Client maximized
	if store.IsMaximized(store.GetInfo(c.Window.Id)) {
		ws := tr.ClientWorkspace(c)
		if ws.TilingDisabled() {
			return
		}
		log.Debug("Client maximized handler fired [", c.Latest.Class, "]")

		// Update client states
		c.Update()

		// Unmaximize window
		c.UnMaximize()

		// Activate maximized layout
		if !c.IsNew() && ws.ActiveLayout().GetName() != "maximized" {
			tr.Channels.Action <- "layout_maximized"
			store.ActiveWindowSet(store.X, c.Window)
		}
	}
}

func (tr *Tracker) handleMinimizedClient(c *store.Client) {
	if !tr.isTracked(c.Window.Id) {
		return
	}

	// Client minimized
	if store.IsMinimized(store.GetInfo(c.Window.Id)) {
		ws := tr.ClientWorkspace(c)
		if ws.TilingDisabled() {
			return
		}
		log.Debug("Client minimized handler fired [", c.Latest.Class, "]")

		// Untrack client
		tr.untrackWindow(c.Window.Id)
	}
}

func (tr *Tracker) handleResizeClient(c *store.Client) {
	ws := tr.ClientWorkspace(c)
	if ws.TilingDisabled() || !tr.isTracked(c.Window.Id) || store.IsMaximized(store.GetInfo(c.Window.Id)) {
		return
	}

	// Previous dimensions
	pGeom := c.Latest.Dimensions.Geometry
	px, py, pw, ph := pGeom.Pieces()

	// Current dimensions
	cGeom, err := c.Window.Instance.DecorGeometry()
	if err != nil {
		return
	}
	cx, cy, cw, ch := cGeom.Pieces()

	// Check size changes
	resized := cw != pw || ch != ph
	moved := (cx != px || cy != py) && (cw == pw && ch == ph)

	if resized && !moved && !tr.Handlers.MoveClient.Active() {
		pt := store.PointerUpdate(store.X)

		// Set client resize event
		if !c.IsNew() && !tr.Handlers.ResizeClient.Active() {
			tr.Handlers.ResizeClient = &Handler{Dragging: pt.Dragging(500), Source: c}
		}
		log.Debug("Client resize handler fired [", c.Latest.Class, "]")

		if tr.Handlers.ResizeClient.Dragging {

			// Set client resize lock
			if tr.Handlers.ResizeClient.Active() {
				tr.Handlers.ResizeClient.Source.(*store.Client).Lock()
				log.Debug("Client resize handler active [", c.Latest.Class, "]")
			}

			// Update proportions
			dir := &store.Directions{
				Top:    cy != py,
				Right:  cx == px && cw != pw,
				Bottom: cy == py && ch != ph,
				Left:   cx != px,
			}
			ws.ActiveLayout().UpdateProportions(c, dir)
		}

		// Tile workspace
		tr.Tile(ws)
	}
}

func (tr *Tracker) handleMoveClient(c *store.Client) {
	ws := tr.ClientWorkspace(c)
	if !tr.isTracked(c.Window.Id) || store.IsMaximized(store.GetInfo(c.Window.Id)) {
		return
	}

	// Previous dimensions
	pGeom := c.Latest.Dimensions.Geometry
	px, py, pw, ph := pGeom.Pieces()

	// Current dimensions
	cGeom, err := c.Window.Instance.DecorGeometry()
	if err != nil {
		return
	}
	cx, cy, cw, ch := cGeom.Pieces()

	// Check position changes
	moved := cx != px || cy != py
	resized := cw != pw || ch != ph

	if moved && !resized && !tr.Handlers.ResizeClient.Active() {
		pt := store.PointerUpdate(store.X)

		// Set client move event
		if !c.IsNew() && !tr.Handlers.MoveClient.Active() {
			tr.Handlers.MoveClient = &Handler{Dragging: pt.Dragging(500), Source: c}
		}
		log.Debug("Client move handler fired [", c.Latest.Class, "]")

		// Obtain targets based on dragging indicator
		targetPoint := *common.CreatePoint(cx, cy)
		if tr.Handlers.MoveClient.Dragging {
			targetPoint = pt.Position
		}
		targetDesktop := store.Workplace.CurrentDesktop
		targetScreen := store.ScreenGet(targetPoint)

		// Check if target point hovers another client
		tr.Handlers.SwapClient.Reset()
		if co := tr.ClientAt(ws, targetPoint); co != nil && co != c {
			tr.Handlers.SwapClient = &Handler{Source: c, Target: co}
			log.Debug("Client swap handler active [", c.Latest.Class, "-", co.Latest.Class, "]")
		}

		// Check if target point moves to another screen
		tr.Handlers.SwapScreen.Reset()
		if c.Latest.Location.Screen != targetScreen {
			tr.Handlers.SwapScreen = &Handler{Source: c, Target: tr.WorkspaceAt(targetDesktop, targetScreen)}
			log.Debug("Screen swap handler active [", c.Latest.Class, "]")
		}
	}
}

func (tr *Tracker) handleSwapClient(h *Handler) {
	c, target := h.Source.(*store.Client), h.Target.(*store.Client)
	ws := tr.ClientWorkspace(c)
	if !tr.isTracked(c.Window.Id) {
		return
	}
	log.Debug("Client swap handler fired [", c.Latest.Class, "-", target.Latest.Class, "]")

	// Swap clients on same desktop and screen
	mg := ws.ActiveLayout().GetManager()
	mg.SwapClient(c, target)

	// Reset client swapping handler
	h.Reset()

	// Tile workspace
	tr.Tile(ws)
}

func (tr *Tracker) handleWorkspaceChange(h *Handler) {
	c, target := h.Source.(*store.Client), h.Target.(*Workspace)
	if !tr.isTracked(c.Window.Id) {
		return
	}
	log.Debug("Client workspace handler fired [", c.Latest.Class, "]")

	// Remove client from current workspace
	ws := tr.ClientWorkspace(c)
	mg := ws.ActiveLayout().GetManager()
	master := mg.IsMaster(c)
	ws.RemoveClient(c)

	// Tile current workspace
	if ws.TilingEnabled() {
		tr.Tile(ws)
	}

	// Update client desktop and screen
	if !tr.isTrackable(c.Window.Id) {
		return
	}
	c.Update()

	// Add client to new workspace
	ws = tr.ClientWorkspace(c)
	if tr.Handlers.SwapScreen.Active() && target.TilingEnabled() {
		ws = target
	}
	mg = ws.ActiveLayout().GetManager()
	ws.AddClient(c)
	if master {
		mg.MakeMaster(c)
	}

	// Tile new workspace
	if ws.TilingEnabled() {
		tr.Tile(ws)
	} else {
		c.Restore(store.Latest)
	}

	// Reset screen swapping handler
	h.Reset()
}

func (tr *Tracker) onStateUpdate(state string, desktop uint, screen uint) {
	workplaceChanged := store.Workplace.DesktopCount*store.Workplace.ScreenCount != uint(len(tr.Workspaces))
	workspaceChanged := common.IsInList(state, []string{"_NET_CURRENT_DESKTOP"})

	viewportChanged := common.IsInList(state, []string{"_NET_NUMBER_OF_DESKTOPS", "_NET_DESKTOP_LAYOUT", "_NET_DESKTOP_GEOMETRY", "_NET_DESKTOP_VIEWPORT", "_NET_WORKAREA"})
	clientsChanged := common.IsInList(state, []string{"_NET_CLIENT_LIST_STACKING"})
	focusChanged := common.IsInList(state, []string{"_NET_ACTIVE_WINDOW"})

	if workplaceChanged {

		// Reset clients and workspaces
		tr.Reset()
	}

	if workspaceChanged {

		// Update sticky windows
		for _, c := range tr.Clients {
			if store.IsSticky(c.Latest) && c.Latest.Location.Desktop != store.Workplace.CurrentDesktop {
				c.MoveDesktop(^uint32(0))
			}
		}
	}

	if viewportChanged || clientsChanged || focusChanged {

		// Deactivate handlers
		tr.Handlers.ResizeClient.Reset()
		tr.Handlers.MoveClient.Reset()
		tr.Handlers.SwapClient.Reset()
		tr.Handlers.SwapScreen.Reset()

		// Unlock clients
		tr.unlockClients()

		// Update trackable clients
		tr.Update()
	}

	if focusChanged {

		// Write client and workspace cache
		tr.Write()
	}
}

func (tr *Tracker) onPointerUpdate(pointer store.XPointer, desktop uint, screen uint) {
	buttonReleased := !pointer.Pressed()

	// Reset timer
	if tr.Handlers.Timer != nil {
		tr.Handlers.Timer.Stop()
	}

	// Wait on button release
	var t time.Duration = 0
	if buttonReleased {
		t = 50
	}

	// Wait for structure events
	tr.Handlers.Timer = time.AfterFunc(t*time.Millisecond, func() {

		// Window moved to another screen
		if tr.Handlers.SwapScreen.Active() {
			tr.handleWorkspaceChange(tr.Handlers.SwapScreen)
		}

		// Window moved over another window
		if tr.Handlers.SwapClient.Active() {
			tr.handleSwapClient(tr.Handlers.SwapClient)
		}

		// Window moved or resized
		if tr.Handlers.MoveClient.Active() || tr.Handlers.ResizeClient.Active() {
			tr.Handlers.MoveClient.Reset()
			tr.Handlers.ResizeClient.Reset()

			// Unlock clients
			tr.unlockClients()

			// Tile workspace
			if buttonReleased {
				tr.Tile(tr.ActiveWorkspace())
			}
		}
	})
}

func (tr *Tracker) attachHandlers(c *store.Client) {
	c.Window.Instance.Listen(xproto.EventMaskStructureNotify | xproto.EventMaskPropertyChange | xproto.EventMaskFocusChange)

	// Attach structure events
	xevent.ConfigureNotifyFun(func(X *xgbutil.XUtil, ev xevent.ConfigureNotifyEvent) {
		log.Trace("Client structure event [", c.Latest.Class, "]")

		// Handle structure events
		tr.handleResizeClient(c)
		tr.handleMoveClient(c)
		if !tr.Handlers.MoveClient.Active() {
			c.Update()
		}
	}).Connect(store.X, c.Window.Id)

	// Attach property events
	xevent.PropertyNotifyFun(func(X *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		aname, _ := xprop.AtomName(store.X, ev.Atom)
		log.Trace("Client property event ", aname, " [", c.Latest.Class, "]")

		// Handle property events
		if aname == "_NET_WM_STATE" {
			tr.handleMaximizedClient(c)
			tr.handleMinimizedClient(c)
		} else if aname == "_NET_WM_DESKTOP" {
			tr.handleWorkspaceChange(&Handler{Source: c, Target: tr.ActiveWorkspace()})
		}
	}).Connect(store.X, c.Window.Id)
}

func (tr *Tracker) isTracked(w xproto.Window) bool {
	_, ok := tr.Clients[w]
	return ok
}

func (tr *Tracker) isTrackable(w xproto.Window) bool {
	info := store.GetInfo(w)
	return !store.IsSpecial(info) && !store.IsIgnored(info)
}

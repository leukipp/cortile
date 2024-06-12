package store

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jezek/xgb/randr"
	"github.com/jezek/xgb/xproto"

	"github.com/jezek/xgbutil"
	"github.com/jezek/xgbutil/ewmh"
	"github.com/jezek/xgbutil/xevent"
	"github.com/jezek/xgbutil/xprop"
	"github.com/jezek/xgbutil/xrect"
	"github.com/jezek/xgbutil/xwindow"

	"github.com/leukipp/cortile/v2/common"

	log "github.com/sirupsen/logrus"
)

var (
	X         *xgbutil.XUtil // X connection
	Workplace *XWorkplace    // X workplace
	Pointer   *XPointer      // X pointer
	Windows   *XWindows      // X windows
)

type XWorkplace struct {
	DeskCount     uint      // Number of desktops
	ScreenCount   uint      // Number of screens
	CurrentDesk   uint      // Current desktop number
	CurrentScreen uint      // Current screen number
	Displays      XDisplays // Physical connected displays
}

type XDisplays struct {
	Name     string    // Unique heads name (display summary)
	Screens  []XHead   // Screen dimensions (full display size)
	Desktops []XHead   // Desktop dimensions (desktop without panels)
	Corners  []*Corner // Display corners (for pointer events)
}

type XHead struct {
	Id       uint32          // Head output id (display id)
	Name     string          // Head output name (display name)
	Primary  bool            // Head primary flag (primary display)
	Geometry common.Geometry // Head dimensions (x/y/width/height)
}

type XPointer struct {
	Drag     XDrag        // Pointer device drag states
	Button   XButton      // Pointer device button states
	Position common.Point // Pointer position coordinates
}

func (p *XPointer) Dragging(dt time.Duration) bool {
	return p.Drag.Left(dt) || p.Drag.Middle(dt) || p.Drag.Right(dt)
}

func (p *XPointer) Pressed() bool {
	return p.Button.Left || p.Button.Middle || p.Button.Right
}

func (p *XPointer) Press() {
	p.Button = XButton{true, true, true}
}

type XDrag struct {
	LeftTime   int64 // Pointer left last drag time
	MiddleTime int64 // Pointer middle last drag time
	RightTime  int64 // Pointer right last drag time
}

func (d *XDrag) Left(dt time.Duration) bool {
	return time.Since(time.UnixMilli(d.LeftTime)) < dt*time.Millisecond
}

func (d *XDrag) Middle(dt time.Duration) bool {
	return time.Since(time.UnixMilli(d.MiddleTime)) < dt*time.Millisecond
}

func (d *XDrag) Right(dt time.Duration) bool {
	return time.Since(time.UnixMilli(d.RightTime)) < dt*time.Millisecond
}

type XButton struct {
	Left   bool // Pointer left click
	Middle bool // Pointer middle click
	Right  bool // Pointer right click
}

type XWindows struct {
	Active  XWindow   // Current active window
	Stacked []XWindow // List of stacked windows
}

type XWindow struct {
	Id       xproto.Window   // Window object id
	Created  int64           // Internal creation timestamp
	Instance *xwindow.Window `json:"-"` // Window object instance
}

func CreateXWindow(w xproto.Window) *XWindow {
	return &XWindow{
		Id:       w,
		Created:  time.Now().UnixMilli(),
		Instance: xwindow.New(X, w),
	}
}

var (
	stateCallbacksFun   []func(string, uint, uint)   // State events callback functions
	pointerCallbacksFun []func(XPointer, uint, uint) // Pointer events callback functions
)

func InitRoot() {

	// Connect to X server
	if !Connected() {
		log.Fatal("Connection to X server failed: exit")
	}

	// Init pointer
	Pointer = PointerGet(X)

	// Init windows
	Windows = &XWindows{}
	Windows.Active = ActiveWindowGet(X)
	Windows.Stacked = ClientListStackingGet(X)

	// Init workplace
	Workplace = &XWorkplace{}
	Workplace.Displays = DisplaysGet(X)
	Workplace.DeskCount = NumberOfDesktopsGet(X)
	Workplace.ScreenCount = uint(len(Workplace.Displays.Screens))
	Workplace.CurrentDesk = CurrentDesktopGet(X)
	Workplace.CurrentScreen = ScreenNumGet(Pointer.Position)

	// Attach root events
	root := CreateXWindow(X.RootWin())
	root.Instance.Listen(xproto.EventMaskSubstructureNotify | xproto.EventMaskPropertyChange)
	xevent.PropertyNotifyFun(StateUpdate).Connect(X, root.Id)
}

func Connected() bool {
	var err error
	var connected bool

	// Retry to connect
	retry := 10
	for i := 0; i <= retry && !connected; i++ {
		if i > 0 {
			log.Warn("Retry in 1 second (", i, "/", retry, ")...")
			time.Sleep(1000 * time.Millisecond)
		}

		// Connect to X server
		X, err = xgbutil.NewConn()
		if err != nil {
			log.Error("Connection to X server failed: ", err)
			continue
		}

		// Check EWMH compliance
		wm, err := ewmh.GetEwmhWM(X)
		if err != nil {
			log.Error("Window manager is not EWMH compliant: ", err)
			continue
		}

		// Validate ROOT properties
		_, err = ewmh.ClientListStackingGet(X)
		if err != nil {
			log.Error("Error retrieving ROOT properties: ", err)
			continue
		}

		// Connection to X established
		log.Info("Connected to X server on ", common.Process.Host, " [", common.Process.System, ", ", wm, "]")
		randr.Init(X.Conn())
		connected = true
	}

	return connected
}

func NumberOfDesktopsGet(X *xgbutil.XUtil) uint {
	deskCount, err := ewmh.NumberOfDesktopsGet(X)

	// Validate number of desktops
	if err != nil {
		log.Error("Error retrieving number of desktops: ", err)
		return Workplace.DeskCount
	}

	return deskCount
}

func CurrentDesktopGet(X *xgbutil.XUtil) uint {
	currentDesk, err := ewmh.CurrentDesktopGet(X)

	// Validate current desktop
	if err != nil {
		log.Error("Error retrieving current desktop: ", err)
		return Workplace.CurrentDesk
	}

	return currentDesk
}

func CurrentDesktopSet(X *xgbutil.XUtil, deskNum uint) {
	ewmh.CurrentDesktopSet(X, deskNum)
	ewmh.ClientEvent(X, X.RootWin(), "_NET_CURRENT_DESKTOP", int(deskNum), int(0))
	Workplace.CurrentDesk = deskNum
}

func ActiveWindowGet(X *xgbutil.XUtil) XWindow {
	active, err := ewmh.ActiveWindowGet(X)

	// Validate active window
	if err != nil {
		log.Error("Error retrieving active window: ", err)
		return Windows.Active
	}

	return *CreateXWindow(active)
}

func ActiveWindowSet(X *xgbutil.XUtil, w *XWindow) {
	ewmh.ActiveWindowSet(X, w.Id)
	ewmh.ClientEvent(X, w.Id, "_NET_ACTIVE_WINDOW", int(2), int(0), int(0))
	Windows.Active = *CreateXWindow(w.Id)
}

func ClientListStackingGet(X *xgbutil.XUtil) []XWindow {
	clients, err := ewmh.ClientListStackingGet(X)

	// Validate client list
	if err != nil {
		log.Error("Error retrieving client list: ", err)
		return Windows.Stacked
	}

	// Create windows
	windows := []XWindow{}
	for _, w := range clients {
		windows = append(windows, *CreateXWindow(w))
	}

	return windows
}

func DisplaysGet(X *xgbutil.XUtil) XDisplays {
	var name string

	// Get geometry of root window
	root := CreateXWindow(X.RootWin())
	geom, err := root.Instance.Geometry()
	if err != nil {
		log.Fatal("Error retrieving root geometry: ", err)
	}

	// Get physical heads
	screens := PhysicalHeadsGet(X)
	desktops := PhysicalHeadsGet(X)

	// Get heads name
	for _, screen := range screens {
		x, y, w, h := screen.Geometry.Pieces()
		name += fmt.Sprintf("%s-%d-%d-%d-%d-%d-", screen.Name, screen.Id, x, y, w, h)
	}
	name = strings.Trim(name, "-")

	// Get desktop rects
	rects := []xrect.Rect{}
	for _, desktop := range desktops {
		rects = append(rects, desktop.Geometry.Rect())
	}

	// Get margins of desktop panels
	for _, w := range Windows.Stacked {
		strut, err := ewmh.WmStrutPartialGet(X, w.Id)
		if err != nil {
			continue
		}

		// Apply struts to rectangles in place
		xrect.ApplyStrut(rects, uint(geom.Width()), uint(geom.Height()),
			strut.Left, strut.Right, strut.Top, strut.Bottom,
			strut.LeftStartY, strut.LeftEndY, strut.RightStartY, strut.RightEndY,
			strut.TopStartX, strut.TopEndX, strut.BottomStartX, strut.BottomEndX,
		)
	}

	// Update desktop geometry
	for i := range desktops {
		desktops[i].Geometry = *common.CreateGeometry(rects[i])
	}

	// Create display heads
	heads := XDisplays{Name: name}
	heads.Screens = screens
	heads.Desktops = desktops
	heads.Corners = CreateCorners(screens)

	// Update screen count
	Workplace.ScreenCount = uint(len(heads.Screens))

	log.Info("Screens ", heads.Screens)
	log.Info("Desktops ", heads.Desktops)

	return heads
}

func PhysicalHeadsGet(X *xgbutil.XUtil) []XHead {

	// Get screen resources
	resources, err := randr.GetScreenResources(X.Conn(), X.RootWin()).Reply()
	if err != nil {
		log.Fatal("Error retrieving screen resources: ", err)
	}

	// Get primary output
	primary, err := randr.GetOutputPrimary(X.Conn(), X.RootWin()).Reply()
	if err != nil {
		log.Fatal("Error retrieving primary screen: ", err)
	}
	hasPrimary := false

	// Get output heads
	heads := []XHead{}
	biggest := XHead{}
	for _, output := range resources.Outputs {
		oinfo, err := randr.GetOutputInfo(X.Conn(), output, 0).Reply()
		if err != nil {
			log.Fatal("Error retrieving screen information: ", err)
		}

		// Ignored screens (disconnected or off)
		if oinfo.Connection != randr.ConnectionConnected || oinfo.Crtc == 0 {
			continue
		}

		// Get crtc information (cathode ray tube controller)
		cinfo, err := randr.GetCrtcInfo(X.Conn(), oinfo.Crtc, 0).Reply()
		if err != nil {
			log.Fatal("Error retrieving screen crtc information: ", err)
		}

		// Append output heads
		head := XHead{
			Id:      uint32(output),
			Name:    string(oinfo.Name),
			Primary: primary != nil && output == primary.Output,
			Geometry: common.Geometry{
				X:      int(cinfo.X),
				Y:      int(cinfo.Y),
				Width:  int(cinfo.Width),
				Height: int(cinfo.Height),
			},
		}
		heads = append(heads, head)

		// Set helper variables
		hasPrimary = head.Primary || hasPrimary
		if head.Geometry.Width*head.Geometry.Height > biggest.Geometry.Width*biggest.Geometry.Height {
			biggest = head
		}
	}

	// Set fallback primary output
	if !hasPrimary {
		for i, head := range heads {
			if head.Id == biggest.Id {
				heads[i].Primary = true
			}
		}
	}

	// Sort output heads
	sort.Slice(heads, func(i, j int) bool {
		return heads[i].Geometry.X < heads[j].Geometry.X
	})

	return heads
}

func PointerGet(X *xgbutil.XUtil) *XPointer {

	// Get current pointer position and button states
	p, err := xproto.QueryPointer(X.Conn(), X.RootWin()).Reply()
	if err != nil {
		log.Warn("Error retrieving pointer position: ", err)
		return Pointer
	}

	return &XPointer{
		Drag: XDrag{},
		Button: XButton{
			Left:   p.Mask&xproto.ButtonMask1 == xproto.ButtonMask1,
			Middle: p.Mask&xproto.ButtonMask2 == xproto.ButtonMask2,
			Right:  p.Mask&xproto.ButtonMask3 == xproto.ButtonMask3,
		},
		Position: common.Point{
			X: int(p.RootX),
			Y: int(p.RootY),
		},
	}
}

func ScreenNumGet(p common.Point) uint {

	// Check if point is inside screen rectangle
	for screenNum, screen := range Workplace.Displays.Screens {
		if common.IsInsideRect(p, screen.Geometry) {
			return uint(screenNum)
		}
	}

	return 0
}

func ScreenGeometry(screenNum uint) *common.Geometry {
	if int(screenNum) >= len(Workplace.Displays.Screens) {
		return &common.Geometry{}
	}
	screen := Workplace.Displays.Screens[screenNum]

	// Get screen geometry
	return &screen.Geometry
}

func DesktopGeometry(screenNum uint) *common.Geometry {
	if int(screenNum) >= len(Workplace.Displays.Desktops) {
		return &common.Geometry{}
	}
	desktop := Workplace.Displays.Desktops[screenNum]

	// Get desktop geometry
	x, y, w, h := desktop.Geometry.Pieces()

	// Add desktop margin
	margin := common.Config.EdgeMargin
	if desktop.Primary && len(common.Config.EdgeMarginPrimary) > 0 {
		margin = common.Config.EdgeMarginPrimary
	}
	if len(margin) == 4 {
		x += margin[3]
		y += margin[0]
		w -= margin[1] + margin[3]
		h -= margin[2] + margin[0]
	}

	return &common.Geometry{
		X:      x,
		Y:      y,
		Width:  w,
		Height: h,
	}
}

func PointerUpdate(X *xgbutil.XUtil) *XPointer {
	previous := XPointer{XDrag{}, XButton{}, common.Point{}}
	if Pointer != nil {
		previous = *Pointer
	}

	// Update current pointer
	Pointer = PointerGet(X)

	// Update current screen
	Workplace.CurrentScreen = ScreenNumGet(Pointer.Position)

	// Update pointer left button drag
	Pointer.Drag.LeftTime = previous.Drag.LeftTime
	if Pointer.Button.Left {
		Pointer.Drag.LeftTime = time.Now().UnixMilli()
	}

	// Update pointer middle button drag
	Pointer.Drag.MiddleTime = previous.Drag.MiddleTime
	if Pointer.Button.Middle {
		Pointer.Drag.MiddleTime = time.Now().UnixMilli()
	}

	// Update pointer right button drag
	Pointer.Drag.RightTime = previous.Drag.RightTime
	if Pointer.Button.Right {
		Pointer.Drag.RightTime = time.Now().UnixMilli()
	}

	// Pointer callbacks
	if previous.Button != Pointer.Button {
		pointerCallbacks(*Pointer, Workplace.CurrentDesk, Workplace.CurrentScreen)
	}

	return Pointer
}

func StateUpdate(X *xgbutil.XUtil, e xevent.PropertyNotifyEvent) {

	// Obtain atom name from property event
	aname, err := xprop.AtomName(X, e.Atom)
	if err != nil {
		log.Warn("Error retrieving atom name: ", err)
		return
	}

	// Update common state variables
	if common.IsInList(aname, []string{"_NET_NUMBER_OF_DESKTOPS"}) {
		Workplace.DeskCount = NumberOfDesktopsGet(X)
	} else if common.IsInList(aname, []string{"_NET_CURRENT_DESKTOP"}) {
		Workplace.CurrentDesk = CurrentDesktopGet(X)
	} else if common.IsInList(aname, []string{"_NET_DESKTOP_LAYOUT", "_NET_DESKTOP_GEOMETRY", "_NET_DESKTOP_VIEWPORT", "_NET_WORKAREA"}) {
		Workplace.Displays = DisplaysGet(X)
	} else if common.IsInList(aname, []string{"_NET_CLIENT_LIST_STACKING"}) {
		Windows.Stacked = ClientListStackingGet(X)
	} else if common.IsInList(aname, []string{"_NET_ACTIVE_WINDOW"}) {
		Windows.Active = ActiveWindowGet(X)
	}
	stateCallbacks(aname, Workplace.CurrentDesk, Workplace.CurrentScreen)
}

func OnPointerUpdate(fun func(XPointer, uint, uint)) {
	pointerCallbacksFun = append(pointerCallbacksFun, fun)
}

func OnStateUpdate(fun func(string, uint, uint)) {
	stateCallbacksFun = append(stateCallbacksFun, fun)
}

func pointerCallbacks(pointer XPointer, desk uint, screen uint) {
	log.Info("Pointer event ", pointer.Button)

	for _, fun := range pointerCallbacksFun {
		fun(pointer, desk, screen)
	}
}

func stateCallbacks(state string, desk uint, screen uint) {
	log.Info("State event ", state)

	for _, fun := range stateCallbacksFun {
		fun(state, desk, screen)
	}
}

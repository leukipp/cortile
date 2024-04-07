package store

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jezek/xgb/randr"
	"github.com/jezek/xgb/render"
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
	Pointer   *XPointer      // X pointer
	Windows   *XWindows      // X windows
	Workplace *XWorkplace    // X workplace
)

var (
	pointerCallbacksFun []func(uint16, uint, uint) // Pointer events callback functions
	stateCallbacksFun   []func(string, uint, uint) // State events callback functions
)

type XPointer struct {
	Button   uint16          // Pointer device button states
	Position render.Pointfix // Pointer position coordinates
}

type XWindows struct {
	Active  xproto.Window   // Current active window
	Stacked []xproto.Window // List of stacked windows
}

type XWorkplace struct {
	Displays      XHeads // Physical connected displays
	DeskCount     uint   // Number of desktops
	ScreenCount   uint   // Number of screens
	CurrentDesk   uint   // Current desktop number
	CurrentScreen uint   // Current screen number
}

type XHeads struct {
	Name     string    // Unique heads name (display summary)
	Screens  []*XHead  // Screen dimensions (full display size)
	Desktops []*XHead  // Desktop dimensions (desktop without panels)
	Corners  []*Corner // Display corners (for pointer events)
}

type XHead struct {
	Id         uint32 // Head output id (display id)
	Name       string // Head output name (display name)
	Primary    bool   // Head primary flag (primary display)
	xrect.Rect        // Head dimensions (x/y/width/height)
}

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
	root := xwindow.New(X, X.RootWin())
	root.Listen(xproto.EventMaskSubstructureNotify | xproto.EventMaskPropertyChange)
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
		log.Info("Connected to X server [", wm, "]")
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

func ActiveWindowGet(X *xgbutil.XUtil) xproto.Window {
	activeWindow, err := ewmh.ActiveWindowGet(X)

	// Validate active window
	if err != nil {
		log.Error("Error retrieving active window: ", err)
		return Windows.Active
	}

	return activeWindow
}

func ClientListStackingGet(X *xgbutil.XUtil) []xproto.Window {
	windows, err := ewmh.ClientListStackingGet(X)

	// Validate client list
	if err != nil {
		log.Error("Error retrieving client list: ", err)
		return Windows.Stacked
	}

	return windows
}

func DisplaysGet(X *xgbutil.XUtil) XHeads {
	var name string

	// Get geometry of root window
	rGeom, err := xwindow.New(X, X.RootWin()).Geometry()
	if err != nil {
		log.Fatal("Error retrieving root geometry: ", err)
	}

	// Get physical heads
	screens := PhysicalHeadsGet(X)
	desktops := PhysicalHeadsGet(X)

	// Get heads name
	for _, screen := range screens {
		x, y, w, h := screen.Rect.Pieces()
		name += fmt.Sprintf("%s-%d-%d-%d-%d-%d-", screen.Name, screen.Id, x, y, w, h)
	}
	name = strings.Trim(name, "-")

	// Get desktop rects
	dRects := []xrect.Rect{}
	for _, desktop := range desktops {
		dRects = append(dRects, desktop.Rect)
	}

	// Account for desktop panels
	for _, win := range Windows.Stacked {
		strut, err := ewmh.WmStrutPartialGet(X, win)
		if err != nil {
			continue
		}

		// Apply in place struts to desktop
		_, _, w, h := rGeom.Pieces()
		xrect.ApplyStrut(dRects, uint(w), uint(h),
			strut.Left, strut.Right, strut.Top, strut.Bottom,
			strut.LeftStartY, strut.LeftEndY, strut.RightStartY, strut.RightEndY,
			strut.TopStartX, strut.TopEndX, strut.BottomStartX, strut.BottomEndX,
		)
	}

	// Create display heads
	heads := XHeads{Name: name}
	heads.Screens = screens
	heads.Desktops = desktops
	heads.Corners = CreateCorners(screens)

	// Update screen count
	Workplace.ScreenCount = uint(len(heads.Screens))

	log.Info("Screens ", heads.Screens)
	log.Info("Desktops ", heads.Desktops)
	log.Info("Corners ", heads.Corners)

	return heads
}

func PhysicalHeadsGet(X *xgbutil.XUtil) []*XHead {

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
	heads := []*XHead{}
	biggest := XHead{Rect: xrect.New(0, 0, 0, 0)}
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
			Rect: xrect.New(
				int(cinfo.X),
				int(cinfo.Y),
				int(cinfo.Width),
				int(cinfo.Height),
			),
		}
		heads = append(heads, &head)

		// Set helper variables
		hasPrimary = head.Primary || hasPrimary
		if head.Width()*head.Height() > biggest.Rect.Width()*biggest.Rect.Height() {
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
		return heads[i].X() < heads[j].X()
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
		Button: p.Mask&xproto.ButtonMask1 | p.Mask&xproto.ButtonMask2 | p.Mask&xproto.ButtonMask3,
		Position: render.Pointfix{
			X: render.Fixed(p.RootX),
			Y: render.Fixed(p.RootY),
		},
	}
}

func ScreenNumGet(p render.Pointfix) uint {

	// Check if point is inside screen rectangle
	for screenNum, rect := range Workplace.Displays.Screens {
		if common.IsInsideRect(p, rect) {
			return uint(screenNum)
		}
	}

	return 0
}

func DesktopDimensions(screenNum uint) (x, y, w, h int) {
	if int(screenNum) >= len(Workplace.Displays.Desktops) {
		return
	}
	desktop := Workplace.Displays.Desktops[screenNum]

	// Get desktop dimensions
	x, y, w, h = desktop.Pieces()

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

	return
}

func PointerUpdate(X *xgbutil.XUtil) *XPointer {

	// Update current pointer
	previousButton := uint16(0)
	if Pointer != nil {
		previousButton = Pointer.Button
	}
	Pointer = PointerGet(X)

	// Update current screen
	Workplace.CurrentScreen = ScreenNumGet(Pointer.Position)

	// Pointer callbacks
	if previousButton != Pointer.Button {
		pointerCallbacks(Pointer.Button, Workplace.CurrentDesk, Workplace.CurrentScreen)
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

func OnPointerUpdate(fun func(uint16, uint, uint)) {
	pointerCallbacksFun = append(pointerCallbacksFun, fun)
}

func OnStateUpdate(fun func(string, uint, uint)) {
	stateCallbacksFun = append(stateCallbacksFun, fun)
}

func pointerCallbacks(button uint16, desk uint, screen uint) {
	log.Info("Pointer event ", button)

	for _, fun := range pointerCallbacksFun {
		fun(button, desk, screen)
	}
}

func stateCallbacks(state string, desk uint, screen uint) {
	log.Info("State event ", state)

	for _, fun := range stateCallbacksFun {
		fun(state, desk, screen)
	}
}

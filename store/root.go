package store

import (
	"time"

	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/leukipp/cortile/common"

	log "github.com/sirupsen/logrus"
)

var (
	X              *xgbutil.XUtil  // X connection object
	DeskCount      uint            // Number of desktops
	ScreenCount    uint            // Number of screens
	CurrentDesk    uint            // Current desktop number
	CurrentScreen  uint            // Current screen number
	CurrentPointer *common.Pointer // Pointer position
	ActiveWindow   xproto.Window   // Current active window
	Windows        []xproto.Window // List of client windows
	Displays       Heads           // Physical connected displays
	Corners        []*Corner       // Corners for pointer events
)

var (
	pointerCallbacksFun []func(uint16) // Pointer events callback functions
	stateCallbacksFun   []func(string) // State events callback functions
)

type Heads struct {
	Screens  []Head // Screen dimensions (full display size)
	Desktops []Head // Desktop dimensions (desktop without panels)
}

type Head struct {
	Id         uint32 // Head output id (display id)
	Name       string // Head output name (display name)
	Primary    bool   // Head primary flag (primary display)
	xrect.Rect        // Head dimensions (x/y/width/height)
}

func InitRoot() {

	// Connect to X server
	X = Connect()

	// Init root properties
	DeskCount = NumberOfDesktopsGet(X)
	CurrentDesk = CurrentDesktopGet(X)
	ActiveWindow = ActiveWindowGet(X)
	Windows = ClientListStackingGet(X)
	Displays = DisplaysGet(X)
	Corners = CreateCorners()

	// Attach root events
	root := xwindow.New(X, X.RootWin())
	root.Listen(xproto.EventMaskSubstructureNotify | xproto.EventMaskPropertyChange)
	xevent.PropertyNotifyFun(StateUpdate).Connect(X, root.Id)
}

func Connect() *xgbutil.XUtil {
	var err error

	// Retry to connect
	for i := 0; i < 10; i++ {
		if i > 0 {
			log.Warn("Retry in 1 second...")
			time.Sleep(1000 * time.Millisecond)
		}

		// Connect to X server
		X, err = xgbutil.NewConn()
		if err != nil {
			log.Error("Connection to X server failed ", err)
			continue
		}

		// Check EWMH compliance
		wm, err := ewmh.GetEwmhWM(X)
		if err != nil {
			log.Error("Window manager is not EWMH compliant ", err)
			continue
		}

		// Validate ROOT properties
		_, err = ewmh.ClientListStackingGet(X)
		if err != nil {
			log.Error("Error retrieving ROOT properties ", err)
			continue
		}

		// Connection established
		log.Info("Connected to X server [", wm, "]")
		randr.Init(X.Conn())

		break
	}

	return X
}

func NumberOfDesktopsGet(X *xgbutil.XUtil) uint {
	deskCount, err := ewmh.NumberOfDesktopsGet(X)

	// Validate number of desktops
	if err != nil {
		log.Error("Error retrieving number of desktops ", err)
		return DeskCount
	}

	return deskCount
}

func CurrentDesktopGet(X *xgbutil.XUtil) uint {
	currentDesk, err := ewmh.CurrentDesktopGet(X)

	// Validate current desktop
	if err != nil {
		log.Error("Error retrieving current desktop ", err)
		return CurrentDesk
	}

	return currentDesk
}

func ActiveWindowGet(X *xgbutil.XUtil) xproto.Window {
	activeWindow, err := ewmh.ActiveWindowGet(X)

	// Validate active window
	if err != nil {
		log.Error("Error retrieving active window ", err)
		return ActiveWindow
	}

	return activeWindow
}

func ClientListStackingGet(X *xgbutil.XUtil) []xproto.Window {
	windows, err := ewmh.ClientListStackingGet(X)

	// Validate client list
	if err != nil {
		log.Error("Error retrieving client list ", err)
		return Windows
	}

	return windows
}

func DisplaysGet(X *xgbutil.XUtil) Heads {

	// Get geometry of root window
	rGeom, err := xwindow.New(X, X.RootWin()).Geometry()
	if err != nil {
		log.Fatal("Error retrieving root geometry ", err)
	}

	// Get physical heads
	screens := PhysicalHeadsGet(X)
	desktops := PhysicalHeadsGet(X)

	// Get bounding rects
	rects := []xrect.Rect{}
	for _, desktop := range desktops {
		rects = append(rects, desktop.Rect)
	}

	// Adjust desktop geometry
	for _, win := range Windows {
		strut, err := ewmh.WmStrutPartialGet(X, win)
		if err != nil {
			continue
		}

		// Apply in place struts to desktop
		xrect.ApplyStrut(rects, uint(rGeom.Width()), uint(rGeom.Height()),
			strut.Left, strut.Right, strut.Top, strut.Bottom,
			strut.LeftStartY, strut.LeftEndY, strut.RightStartY, strut.RightEndY,
			strut.TopStartX, strut.TopEndX, strut.BottomStartX, strut.BottomEndX,
		)
	}

	// Update screen count
	ScreenCount = uint(len(screens))

	log.Info("Screens ", screens)
	log.Info("Desktops ", desktops)

	return Heads{Screens: screens, Desktops: desktops}
}

func PhysicalHeadsGet(X *xgbutil.XUtil) []Head {

	// Get screen resources
	resources, err := randr.GetScreenResources(X.Conn(), X.RootWin()).Reply()
	if err != nil {
		log.Fatal("Error retrieving screen resources ", err)
	}

	// Get primary output
	primary, err := randr.GetOutputPrimary(X.Conn(), X.RootWin()).Reply()
	if err != nil {
		log.Fatal("Error retrieving primary screen ", err)
	}
	hasPrimary := false

	// Get output heads
	heads := []Head{}
	biggest := Head{Rect: xrect.New(0, 0, 0, 0)}
	for _, output := range resources.Outputs {
		oinfo, err := randr.GetOutputInfo(X.Conn(), output, 0).Reply()
		if err != nil {
			log.Fatal("Error retrieving screen information ", err)
		}

		// Ignored screens (disconnected or off)
		if oinfo.Connection != randr.ConnectionConnected || oinfo.Crtc == 0 {
			continue
		}

		// Get crtc information (cathode ray tube controller)
		cinfo, err := randr.GetCrtcInfo(X.Conn(), oinfo.Crtc, 0).Reply()
		if err != nil {
			log.Fatal("Error retrieving screen crtc information ", err)
		}

		// Append output heads
		head := Head{
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
		heads = append(heads, head)

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

	return heads
}

func PointerGet(X *xgbutil.XUtil) *common.Pointer {

	// Get current pointer position and button states
	p, err := xproto.QueryPointer(X.Conn(), X.RootWin()).Reply()
	if err != nil {
		log.Warn("Error retrieving pointer position ", err)
		return CurrentPointer
	}

	return &common.Pointer{
		X:      p.RootX,
		Y:      p.RootY,
		Button: p.Mask&xproto.ButtonMask1 | p.Mask&xproto.ButtonMask2 | p.Mask&xproto.ButtonMask3,
	}
}

func ScreenNumGet(p *common.Pointer) uint {

	// Check if point is inside screen rectangle
	for screenNum, rect := range Displays.Screens {
		if common.IsInsideRect(p, rect) {
			return uint(screenNum)
		}
	}

	return 0
}

func DesktopDimensions(screenNum uint) (x, y, w, h int) {
	if int(screenNum) >= len(Displays.Desktops) {
		return
	}
	desktop := Displays.Desktops[screenNum]

	// Get desktop dimensions
	x, y, w, h = desktop.Pieces()

	// Add desktop margin
	if desktop.Primary {
		x += common.Config.EdgeMargin[3]
		y += common.Config.EdgeMargin[0]
		w -= common.Config.EdgeMargin[1] + common.Config.EdgeMargin[3]
		h -= common.Config.EdgeMargin[2] + common.Config.EdgeMargin[0]
	}

	return
}

func PointerUpdate(X *xgbutil.XUtil) {

	// Update current pointer
	previousButton := uint16(0)
	if CurrentPointer != nil {
		previousButton = CurrentPointer.Button
	}
	CurrentPointer = PointerGet(X)
	if previousButton != CurrentPointer.Button {
		pointerCallbacks(CurrentPointer.Button)
	}

	// Update current screen
	CurrentScreen = ScreenNumGet(CurrentPointer)
}

func StateUpdate(X *xgbutil.XUtil, e xevent.PropertyNotifyEvent) {

	// Obtain atom name from property event
	aname, err := xprop.AtomName(X, e.Atom)
	if err != nil {
		log.Warn("Error retrieving atom name ", err)
		return
	}

	// Update common state variables
	if common.IsInList(aname, []string{"_NET_NUMBER_OF_DESKTOPS"}) {
		DeskCount = NumberOfDesktopsGet(X)
		stateCallbacks(aname)
	} else if common.IsInList(aname, []string{"_NET_CURRENT_DESKTOP"}) {
		CurrentDesk = CurrentDesktopGet(X)
		stateCallbacks(aname)
	} else if common.IsInList(aname, []string{"_NET_DESKTOP_LAYOUT", "_NET_DESKTOP_GEOMETRY", "_NET_DESKTOP_VIEWPORT", "_NET_WORKAREA"}) {
		Displays = DisplaysGet(X)
		Corners = CreateCorners()
		stateCallbacks(aname)
	} else if common.IsInList(aname, []string{"_NET_CLIENT_LIST_STACKING"}) {
		Windows = ClientListStackingGet(X)
		stateCallbacks(aname)
	} else if common.IsInList(aname, []string{"_NET_ACTIVE_WINDOW"}) {
		ActiveWindow = ActiveWindowGet(X)
		stateCallbacks(aname)
	}
}

func OnPointerUpdate(fun func(uint16)) {
	pointerCallbacksFun = append(pointerCallbacksFun, fun)
}

func OnStateUpdate(fun func(string)) {
	stateCallbacksFun = append(stateCallbacksFun, fun)
}

func pointerCallbacks(arg uint16) {
	log.Info("Pointer event ", arg)

	for _, fun := range pointerCallbacksFun {
		fun(arg)
	}
}

func stateCallbacks(arg string) {
	log.Info("State event ", arg)

	for _, fun := range stateCallbacksFun {
		fun(arg)
	}
}

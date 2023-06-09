package store

import (
	"time"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xinerama"
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
	ViewPorts      Head            // Physical monitors
	ActiveWindow   xproto.Window   // Current active window
	Windows        []xproto.Window // List of client windows
	Corners        []*Corner       // Corners for pointer events
)

var (
	pointerCallbacksFun []func(uint16) // Pointer events callback functions
	stateCallbacksFun   []func(string) // State events callback functions
)

type Head struct {
	Screens  xinerama.Heads // Screen size (full monitor size)
	Desktops xinerama.Heads // Desktop size (workarea without panels)
}

func InitRoot() {
	var err error

	X := Connect()
	root := xwindow.New(X, X.RootWin())

	DeskCount, err = ewmh.NumberOfDesktopsGet(X)
	checkFatal(err)

	CurrentDesk, err = ewmh.CurrentDesktopGet(X)
	checkFatal(err)

	ViewPorts, err = ViewPortsGet(X)
	checkFatal(err)

	ActiveWindow, err = ewmh.ActiveWindowGet(X)
	checkFatal(err)

	Windows, err = ewmh.ClientListGet(X)
	checkFatal(err)

	Corners = CreateCorners()

	root.Listen(xproto.EventMaskPropertyChange)
	xevent.PropertyNotifyFun(StateUpdate).Connect(X, X.RootWin())
}

func Connect() *xgbutil.XUtil {
	var err error

	// Connect to X server
	X, err = xgbutil.NewConn()
	checkFatal(err)

	// Check ewmh compliance
	_, err = ewmh.GetEwmhWM(X)
	if err != nil {
		log.Fatal("Window manager is not EWMH compliant ", err)
	}

	// Wait for client list availability
	i, j := 0, 100
	for i < j {
		_, err = ewmh.ClientListStackingGet(X)
		if err == nil {
			break
		}
		i += 1
		time.Sleep(100 * time.Millisecond)
	}

	log.Info("Connected to X server")

	return X
}

func PhysicalHeadsGet(rGeom xrect.Rect) xinerama.Heads {

	// Get the physical heads
	heads := xinerama.Heads{rGeom}
	if X.ExtInitialized("XINERAMA") {
		heads, _ = xinerama.PhysicalHeads(X)
	}

	return heads
}

func ViewPortsGet(X *xgbutil.XUtil) (Head, error) {

	// Get the geometry of the root window
	rGeom, err := xwindow.New(X, X.RootWin()).Geometry()
	checkFatal(err)

	// Get the physical heads
	screens := PhysicalHeadsGet(rGeom)
	desktops := PhysicalHeadsGet(rGeom)

	// Adjust desktops geometry
	clients, err := ewmh.ClientListStackingGet(X)
	for _, id := range clients {
		strut, err := ewmh.WmStrutPartialGet(X, id)
		if err != nil {
			continue
		}

		// Apply in place struts to desktops
		xrect.ApplyStrut(desktops, uint(rGeom.Width()), uint(rGeom.Height()),
			strut.Left, strut.Right, strut.Top, strut.Bottom,
			strut.LeftStartY, strut.LeftEndY,
			strut.RightStartY, strut.RightEndY,
			strut.TopStartX, strut.TopEndX,
			strut.BottomStartX, strut.BottomEndX)
	}

	// Update screen count
	ScreenCount = uint(len(screens))

	log.Info("Screens ", screens)
	log.Info("Desktops ", desktops)

	return Head{Screens: screens, Desktops: desktops}, err
}

func PointerGet(X *xgbutil.XUtil) *common.Pointer {

	// Get current pointer position and button states
	p, err := xproto.QueryPointer(X.Conn(), X.RootWin()).Reply()
	if err != nil {
		log.Warn("Error on pointer update ", err)
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
	for screenNum, rect := range ViewPorts.Screens {
		if common.IsInsideRect(p, rect) {
			return uint(screenNum)
		}
	}

	return 0
}

func DesktopDimensions(screenNum uint) (x, y, w, h int) {
	x, y, w, h = ViewPorts.Desktops[screenNum].Pieces()

	// Add desktop margin
	x += common.Config.EdgeMargin[3]
	y += common.Config.EdgeMargin[0]
	w -= common.Config.EdgeMargin[1] + common.Config.EdgeMargin[3]
	h -= common.Config.EdgeMargin[2] + common.Config.EdgeMargin[0]

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
	var err error

	// Obtain atom name from notify event
	aname, err := xprop.AtomName(X, e.Atom)

	// Update common state variables
	if common.IsInList(aname, []string{"_NET_NUMBER_OF_DESKTOPS"}) {
		DeskCount, err = ewmh.NumberOfDesktopsGet(X)
		stateCallbacks(aname)
	} else if common.IsInList(aname, []string{"_NET_CURRENT_DESKTOP"}) {
		CurrentDesk, err = ewmh.CurrentDesktopGet(X)
		stateCallbacks(aname)
	} else if common.IsInList(aname, []string{"_NET_DESKTOP_LAYOUT", "_NET_DESKTOP_GEOMETRY", "_NET_DESKTOP_VIEWPORT", "_NET_WORKAREA"}) {
		ViewPorts, err = ViewPortsGet(X)
		Corners = CreateCorners()
		stateCallbacks(aname)
	} else if common.IsInList(aname, []string{"_NET_CLIENT_LIST_STACKING"}) {
		Windows, err = ewmh.ClientListStackingGet(X)
		stateCallbacks(aname)
	} else if common.IsInList(aname, []string{"_NET_ACTIVE_WINDOW"}) {
		ActiveWindow, err = ewmh.ActiveWindowGet(X)
		stateCallbacks(aname)
	}

	if err != nil {
		log.Warn("Error on state update ", err)
		return
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

func checkFatal(err error) {
	if err != nil {
		log.Fatal("Error on initialization ", err)
	}
}

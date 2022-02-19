package common

import (
	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xinerama"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	log "github.com/sirupsen/logrus"
)

var (
	X           *xgbutil.XUtil  // X connection object
	DeskCount   uint            // Number of desktop workspaces
	CurrentDesk uint            // Current desktop
	ViewPorts   Head            // Physical monitors
	Stacking    []xproto.Window // List of client windows
	ActiveWin   xproto.Window   // Current active window
	Corners     []Corner        // Corners for pointer events
	//WorkArea   []ewmh.Workarea // Work area on desktop
)

type Head struct {
	Screens  xinerama.Heads // Full screen size
	Desktops xinerama.Heads // Desktop size (without menu offset)
}

// Populate initializes the state variables and registers the callbacks required for keeping them up-to-date.
func Init() {
	var err error

	X, err = xgbutil.NewConn()
	checkFatal(err)
	checkEwmhCompliance()

	DeskCount, err = ewmh.NumberOfDesktopsGet(X)
	checkFatal(err)

	CurrentDesk, err = ewmh.CurrentDesktopGet(X)
	checkFatal(err)

	//WorkArea, err = ewmh.WorkareaGet(X)
	//checkFatal(err)

	ViewPorts, err = ViewPortsGet(X)
	checkError(err)

	Stacking, err = ewmh.ClientListStackingGet(X)
	checkError(err)

	ActiveWin, err = ewmh.ActiveWindowGet(X)
	checkError(err)

	Corners = CreateCorners()

	root := xwindow.New(X, X.RootWin())
	root.Listen(xproto.EventMaskPropertyChange)

	xevent.PropertyNotifyFun(stateUpdate).Connect(X, X.RootWin())
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
	root := xwindow.New(X, X.RootWin())
	rGeom, err := root.Geometry()
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

		// Apply struts to our desktops 'in place'
		xrect.ApplyStrut(desktops, uint(rGeom.Width()), uint(rGeom.Height()),
			strut.Left, strut.Right, strut.Top, strut.Bottom,
			strut.LeftStartY, strut.LeftEndY,
			strut.RightStartY, strut.RightEndY,
			strut.TopStartX, strut.TopEndX,
			strut.BottomStartX, strut.BottomEndX)
	}

	log.Info("Screens ", screens)
	log.Info("Desktops ", desktops)

	return Head{Screens: screens, Desktops: desktops}, err
}

// WorkAreaDimensions returns the dimension of the requested workspace.
func WorkAreaDimensions(num uint) (x, y, w, h int) {
	x, y, w, h = DesktopDimensions()
	return

	// TODO: evaluate workarea vs. desktop
	//wa := WorkArea[num]
	//x, y = wa.X, wa.Y
	//w, h = int(wa.Width), int(wa.Height)
	//return
}

func DesktopDimensions() (x, y, w, h int) {
	for _, d := range ViewPorts.Desktops {
		hx, hy := d.X(), d.Y()
		hw, hh := int(d.Width()), int(d.Height())

		// use biggest head (monitor) as working area
		if hw*hh > w*h {
			x, y = hx, hy
			w, h = hw, hh
		}
	}

	return
}

func ScreenDimensions() (x, y, w, h int) {
	for _, s := range ViewPorts.Screens {
		hx, hy := s.X(), s.Y()
		hw, hh := int(s.Width()), int(s.Height())

		// use biggest head (monitor) as working area
		if hw*hh > w*h {
			x, y = hx, hy
			w, h = hw, hh
		}
	}

	return
}

func checkEwmhCompliance() {
	_, err := ewmh.GetEwmhWM(X)
	if err != nil {
		log.Fatal("Window manager is not EWMH complaint!")
	}
}

func checkFatal(err error) {
	if err != nil {
		log.Fatal("Error populating state ", err)
	}
}

func checkError(err error) {
	if err != nil {
		log.Error("Warning populating state ", err)
	}
}

func stateUpdate(X *xgbutil.XUtil, e xevent.PropertyNotifyEvent) {
	var err error

	aname, _ := xprop.AtomName(X, e.Atom)

	log.Debug("State event ", aname)

	if aname == "_NET_NUMBER_OF_DESKTOPS" {
		DeskCount, err = ewmh.NumberOfDesktopsGet(X)
	} else if aname == "_NET_CURRENT_DESKTOP" {
		CurrentDesk, err = ewmh.CurrentDesktopGet(X)
	} else if aname == "_NET_DESKTOP_VIEWPORT" {
		ViewPorts, err = ViewPortsGet(X)
		Corners = CreateCorners()
	} else if aname == "_NET_CLIENT_LIST_STACKING" {
		Stacking, err = ewmh.ClientListStackingGet(X)
	} else if aname == "_NET_ACTIVE_WINDOW" {
		ActiveWin, err = ewmh.ActiveWindowGet(X)
	}

	if err != nil {
		log.Warn("Warning updating state ", err)
	}
}

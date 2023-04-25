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
	Corners     []*Corner       // Corners for pointer events
	Pointer     Position        // Pointer position
)

type Head struct {
	Screens  xinerama.Heads // Screen size (full monitor size)
	Desktops xinerama.Heads // Desktop size (workarea without panels)
}

type Position struct {
	X int // X position
	Y int // Y position
}

func InitState() {
	var err error

	X, err = xgbutil.NewConn()
	checkFatal(err)
	checkEwmhCompliance()

	DeskCount, err = ewmh.NumberOfDesktopsGet(X)
	checkFatal(err)

	CurrentDesk, err = ewmh.CurrentDesktopGet(X)
	checkFatal(err)

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

		// Apply in place struts to our desktops
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

func DesktopDimensions() (x, y, w, h int) {
	for _, d := range ViewPorts.Desktops {
		hx, hy, hw, hh := d.Pieces()

		// Use biggest head (monitor) as desktop area
		if hw*hh > w*h {
			x, y, w, h = hx, hy, hw, hh
		}
	}

	// Add desktop margin
	x += Config.EdgeMargin[3]
	y += Config.EdgeMargin[0]
	w -= Config.EdgeMargin[1] + Config.EdgeMargin[3]
	h -= Config.EdgeMargin[2] + Config.EdgeMargin[0]

	return
}

func ScreenDimensions() (x, y, w, h int) {
	for _, s := range ViewPorts.Screens {
		hx, hy, hw, hh := s.Pieces()

		// Use biggest head (monitor) as screen area
		if hw*hh > w*h {
			x, y, w, h = hx, hy, hw, hh
		}
	}

	return
}

func stateUpdate(X *xgbutil.XUtil, e xevent.PropertyNotifyEvent) {
	var err error

	aname, _ := xprop.AtomName(X, e.Atom)

	log.Trace("State event ", aname)

	// Update common state variables
	if aname == "_NET_NUMBER_OF_DESKTOPS" {
		DeskCount, err = ewmh.NumberOfDesktopsGet(X)
	} else if aname == "_NET_CURRENT_DESKTOP" {
		CurrentDesk, err = ewmh.CurrentDesktopGet(X)
	} else if aname == "_NET_DESKTOP_VIEWPORT" {
		ViewPorts, err = ViewPortsGet(X)
		Corners = CreateCorners()
	} else if aname == "_NET_WORKAREA" {
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

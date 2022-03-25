package store

import (
	"math"
	"reflect"
	"regexp"
	"strings"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/leukipp/cortile/common"

	log "github.com/sirupsen/logrus"
)

var UNKNOWN = "<UNKNOWN>"

type Client struct {
	Win         *xwindow.Window // X window object
	Desk        uint            // Desktop the client is currently in
	Name        string          // Client window title name
	Class       string          // Client window application name
	CurrentProp Property        // Properties that the client has at the moment
	SavedProp   Property        // Properties that the client had before tiling
}

type Property struct {
	Geom xrect.Rect // Client rectangle geometry
	Deco bool       // Decoration active or not
}

func CreateClient(w xproto.Window) (c *Client) {
	win := xwindow.New(common.X, w)
	class, name, desk, _, _ := GetInfo(w)

	savedGeom, err := win.DecorGeometry()
	if err != nil {
		log.Info(err)
	}

	return &Client{
		Win:   win,
		Desk:  desk,
		Name:  name,
		Class: class,
		CurrentProp: Property{
			Geom: savedGeom,
			Deco: HasDecoration(w),
		},
		SavedProp: Property{
			Geom: savedGeom,
			Deco: HasDecoration(w),
		},
	}
}

func (c *Client) MoveResize(x, y, w, h int) {
	c.Unmaximize()

	dw, dh := c.DecorDimensions()

	// Move window
	err := c.Win.WMMoveResize(x, y, w-dw, h-dh)
	if err != nil {
		log.Warn("Error when moving window [", c.Class, "]")
	}

	// Update stored dimensions
	c.Update()
}

func (c *Client) DecorDimensions() (w int, h int) {

	// Inner dimension
	cGeom, err1 := xwindow.RawGeometry(common.X, xproto.Drawable(c.Win.Id))
	if err1 != nil {
		log.Warn(err1)
		return
	}

	// Outer dimension
	pGeom, err2 := c.Win.DecorGeometry()
	if err2 != nil {
		log.Warn(err2)
		return
	}

	w, h = pGeom.Width()-cGeom.Width(), pGeom.Height()-cGeom.Height()

	return
}

func (c *Client) Update() (success bool) {
	class, name, desk, _, _ := GetInfo(c.Win.Id)
	if class == UNKNOWN {
		return false
	}

	// Set client properties
	c.Class = class
	c.Name = name
	c.Desk = desk

	// Update client geometry
	pGeom, err := c.Win.DecorGeometry()
	if err != nil {
		log.Warn(err)
		return false
	}
	c.CurrentProp.Geom = pGeom

	return true
}

func (c Client) Activate() {
	ewmh.ActiveWindowReq(common.X, c.Win.Id)
}

func (c Client) Unmaximize() {
	ewmh.WmStateReq(common.X, c.Win.Id, 0, "_NET_WM_STATE_MAXIMIZED_VERT")
	ewmh.WmStateReq(common.X, c.Win.Id, 0, "_NET_WM_STATE_MAXIMIZED_HORZ")
}

func (c Client) UnDecorate() {
	motif.WmHintsSet(common.X, c.Win.Id,
		&motif.Hints{
			Flags:      motif.HintDecorations,
			Decoration: motif.DecorationNone,
		})
}

func (c Client) Decorate() {
	if !c.SavedProp.Deco {
		return
	}

	motif.WmHintsSet(common.X, c.Win.Id,
		&motif.Hints{
			Flags:      motif.HintDecorations,
			Decoration: motif.DecorationAll,
		})
}

func (c Client) Restore() {
	c.Decorate()

	// Move window to stored position
	geom := c.SavedProp.Geom
	c.MoveResize(geom.X(), geom.Y(), geom.Width(), geom.Height())

	log.Info("Restoring window position x=", geom.X(), ", y=", geom.Y(), " [", c.Class, "]")
}

func GetInfo(w xproto.Window) (class string, name string, desk uint, states []string, hints *motif.Hints) {
	var err error
	var wmClass *icccm.WmClass

	// Class name
	wmClass, err = icccm.WmClassGet(common.X, w)
	if err != nil {
		log.Trace(err)
		class = UNKNOWN
	} else if wmClass != nil {
		class = wmClass.Class
	}

	// Windows title
	name, err = icccm.WmNameGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		name = UNKNOWN
	}

	// Window desktop
	desk, err = ewmh.WmDesktopGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		desk = math.MaxUint
	}

	// Window states
	states, err = ewmh.WmStateGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		states = []string{}
	}

	// Window hints
	hints, err = motif.WmHintsGet(common.X, w)
	if err != nil {
		hints = &motif.Hints{}
	}

	return class, name, desk, states, hints
}

func HasDecoration(w xproto.Window) bool {
	_, _, _, _, hints := GetInfo(w)
	return motif.Decor(hints)
}

func IsMaximized(w xproto.Window) bool {
	class, name, _, states, _ := GetInfo(w)
	if class == UNKNOWN {
		return false
	}

	// Check maximized windows
	for _, state := range states {
		if strings.Contains(state, "_NET_WM_STATE_MAXIMIZED") {
			log.Info("Ignore maximized window", " [", name, "]")
			return true
		}
	}

	return false
}

func IsHidden(w xproto.Window) bool {
	class, name, _, states, _ := GetInfo(w)
	if class == UNKNOWN {
		return true
	}

	// Check hidden windows
	for _, state := range states {
		if state == "_NET_WM_STATE_HIDDEN" {
			log.Info("Ignore hidden window", " [", name, "]")
			return true
		}
	}

	return false
}

func IsModal(w xproto.Window) bool {
	class, name, _, states, _ := GetInfo(w)
	if class == UNKNOWN {
		return true
	}

	// Check model dialog windows
	for _, state := range states {
		if state == "_NET_WM_STATE_MODAL" {
			log.Info("Ignore modal window", " [", name, "]")
			return true
		}
	}

	return false
}

func IsIgnored(w xproto.Window) bool {
	class, name, _, _, _ := GetInfo(w)
	if class == UNKNOWN {
		return true
	}

	// Check ignored windows
	for _, s := range common.Config.WindowIgnore {
		conf_class := s[0]
		conf_name := s[1]

		reg_class := regexp.MustCompile(strings.ToLower(conf_class))
		reg_name := regexp.MustCompile(strings.ToLower(conf_name))

		// Ignore all windows with this class
		class_match := reg_class.MatchString(strings.ToLower(class))

		// But allow the window with a special name
		name_match := conf_name != "" && reg_name.MatchString(strings.ToLower(name))

		if class_match && !name_match {
			log.Info("Ignore window with ", strings.TrimSpace(strings.Join(s, " ")), " from config [", name, "]")
			return true
		}
	}

	return false
}

func IsInsideViewPort(w xproto.Window) bool {
	class, _, desk, _, _ := GetInfo(w)
	if class == UNKNOWN {
		return false
	}

	// Ignore pinned windows
	if desk > common.DeskCount {
		log.Info("Ignore pinned window [", class, "]")
		return false
	}

	// Window dimensions
	wGeom, err := xwindow.New(common.X, w).DecorGeometry()
	if err != nil {
		log.Warn(err)
		return false
	}

	// Viewport dimensions
	vRect := xrect.New(common.ScreenDimensions())

	// Substract viewport rectangle (r2) from window rectangle (r1)
	sRects := xrect.Subtract(wGeom, vRect)

	// If r1 does not overlap r2, then only one rectangle is returned which is equivalent to r1
	isOutsideViewport := false
	if len(sRects) == 1 {
		isOutsideViewport = reflect.DeepEqual(sRects[0], wGeom)
	}

	if isOutsideViewport {
		log.Info("Ignore window outside viewport [", class, "]")
	}

	return !isOutsideViewport
}

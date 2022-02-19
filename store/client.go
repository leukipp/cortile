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
	Win         *xwindow.Window
	Desk        uint   // Desktop the client is currently in.
	Name        string // Window title name.
	Class       string // Window application name.
	CurrentProp Prop   // Properties that the client has at the moment.
	SavedProp   Prop   // Properties that the client had before it was tiled.
}

type Prop struct {
	Geom       xrect.Rect
	Decoration bool
}

func CreateClient(w xproto.Window) (c Client) {
	win := xwindow.New(common.X, w)
	class, name, desk, _, _ := GetInfo(w)

	savedGeom, err := win.DecorGeometry()
	if err != nil {
		log.Info(err)
	}

	c = Client{
		Win:   win,
		Desk:  desk,
		Name:  name,
		Class: class,
		CurrentProp: Prop{
			Geom:       savedGeom,
			Decoration: HasDecoration(w),
		},
		SavedProp: Prop{
			Geom:       savedGeom,
			Decoration: HasDecoration(w),
		},
	}

	return c
}

func (c *Client) MoveResize(x, y, w, h int) {
	c.Unmaximize()

	dw, dh := c.DecorDimensions()

	err := c.Win.WMMoveResize(x, y, w-dw, h-dh)
	if err != nil {
		log.Warn("Error when moving window [", c.Class, "]")
	}

	c.Update()
}

func (c *Client) DecorDimensions() (w int, h int) {
	cGeom, err1 := xwindow.RawGeometry(common.X, xproto.Drawable(c.Win.Id))
	if err1 != nil {
		log.Warn(err1)
		return
	}

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

	c.Class = class
	c.Name = name
	c.Desk = desk

	pGeom, err := c.Win.DecorGeometry()
	if err != nil {
		log.Warn(err)
		return false
	}
	c.CurrentProp.Geom = pGeom

	return true
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
	if !c.SavedProp.Decoration {
		return
	}

	motif.WmHintsSet(common.X, c.Win.Id,
		&motif.Hints{
			Flags:      motif.HintDecorations,
			Decoration: motif.DecorationAll,
		})
}

// Restore resizes and decorates window to pre-tiling state.
func (c Client) Restore() {
	c.Decorate()

	geom := c.SavedProp.Geom
	c.MoveResize(geom.X(), geom.Y(), geom.Width(), geom.Height())

	log.Info("Restoring window position x=", geom.X(), ", y=", geom.Y(), " [", c.Class, "]")
}

// Activate makes the client the currently active window.
func (c Client) Activate() {
	ewmh.ActiveWindowReq(common.X, c.Win.Id)
}

// Get window info.
func GetInfo(w xproto.Window) (class string, name string, desk uint, states []string, hints *motif.Hints) {
	var err error
	var wmClass *icccm.WmClass

	wmClass, err = icccm.WmClassGet(common.X, w)
	if err != nil {
		log.Trace(err)
		class = UNKNOWN
	} else if wmClass != nil {
		class = wmClass.Class
	}

	name, err = icccm.WmNameGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		name = UNKNOWN
	}

	desk, err = ewmh.WmDesktopGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		desk = math.MaxUint
	}

	states, err = ewmh.WmStateGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		states = []string{}
	}

	hints, err = motif.WmHintsGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		hints = &motif.Hints{}
	}

	return class, name, desk, states, hints
}

// hasDecoration returns true if the window has client decorations.
func HasDecoration(w xproto.Window) bool {
	_, _, _, _, hints := GetInfo(w)
	return motif.Decor(hints)
}

// isMaximized returns true if the window has been maximized.
func IsMaximized(w xproto.Window) bool {
	class, name, _, states, _ := GetInfo(w)
	if class == UNKNOWN {
		return false
	}

	for _, state := range states {
		if strings.Contains(state, "_NET_WM_STATE_MAXIMIZED") {
			log.Info("Ignore maximized window", " [", name, "]")
			return true
		}
	}

	return false
}

// isHidden returns true if the window has been minimized.
func IsHidden(w xproto.Window) bool {
	class, name, _, states, _ := GetInfo(w)
	if class == UNKNOWN {
		return true
	}

	for _, state := range states {
		if state == "_NET_WM_STATE_HIDDEN" {
			log.Info("Ignore hidden window", " [", name, "]")
			return true
		}
	}

	return false
}

// isModal returns true if the window is a modal dialog.
func IsModal(w xproto.Window) bool {
	class, name, _, states, _ := GetInfo(w)
	if class == UNKNOWN {
		return true
	}

	for _, state := range states {
		if state == "_NET_WM_STATE_MODAL" {
			log.Info("Ignore modal window", " [", name, "]")
			return true
		}
	}

	return false
}

// isIgnored returns true if the window is ignored by config.
func IsIgnored(w xproto.Window) bool {
	class, name, _, _, _ := GetInfo(w)
	if class == UNKNOWN {
		return true
	}

	for _, s := range common.Config.WindowsToIgnore {
		conf_class := s[0]
		conf_name := s[1]

		reg_class := regexp.MustCompile(strings.ToLower(conf_class))
		reg_name := regexp.MustCompile(strings.ToLower(conf_name))

		// ignore all windows with this class...
		class_match := reg_class.MatchString(strings.ToLower(class))

		// ...except the window with a special name
		name_match := conf_name != "" && reg_name.MatchString(strings.ToLower(name))

		if class_match && !name_match {
			log.Info("Ignore window with ", strings.TrimSpace(strings.Join(s, " ")), " from config [", name, "]")
			return true
		}
	}

	return false
}

// isInsideViewPort returns true if the window is partially inside viewport.
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
	wx, wy, ww, wh := wGeom.X(), wGeom.Y(), wGeom.Width(), wGeom.Height()
	wRect := xrect.New(wx, wy, ww, wh)

	// Viewport dimensions
	vx, vy, vw, vh := common.ScreenDimensions()
	vRect := xrect.New(vx, vy, vw, vh)

	// Substract viewport rectangle (r2) from window rectangle (r1)
	sRects := xrect.Subtract(wRect, vRect)

	// If r1 does not overlap r2, then only one rectangle is returned and is equivalent to r1
	isOutsideViewport := false
	if len(sRects) == 1 {
		isOutsideViewport = reflect.DeepEqual(sRects[0], wRect)
	}

	if isOutsideViewport {
		log.Info("Ignore window outside viewport [", class, "]")
	}

	return !isOutsideViewport
}

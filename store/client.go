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
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/leukipp/cortile/common"

	log "github.com/sirupsen/logrus"
)

var UNKNOWN = "<UNKNOWN>"

type Client struct {
	Win      *xwindow.Window // X window object
	Latest   Info            // Latest client window information
	Original Info            // Original client window information
}

type Info struct {
	Class    string            // Client window application name
	Name     string            // Client window title name
	Desk     uint              // Client window desktop
	States   []string          // Client window states
	Hints    motif.Hints       // Client window hints
	Extents  ewmh.FrameExtents // Client window extents
	Geometry xrect.Rect        // Client window geometry
}

func CreateClient(w xproto.Window) (c *Client) {
	info := GetInfo(w)
	return &Client{
		Win:      xwindow.New(common.X, w),
		Latest:   info,
		Original: info,
	}
}

func (c *Client) MoveResize(x, y, w, h int) {
	c.Unmaximize()

	// Decoration extents
	extents := c.Latest.Extents

	// Calculate dimensions offsets
	dx, dy := 0, 0
	if c.Latest.Hints.Flags&motif.HintDecorations > 0 {
		dx, dy = extents.Left, extents.Top
	}
	dw, dh := extents.Left+extents.Right, extents.Top+extents.Bottom

	// Move and resize window
	err := ewmh.MoveresizeWindow(c.Win.X, c.Win.Id, x+dx, y+dy, w-dw, h-dh)
	if err != nil {
		log.Warn("Error when moving window [", c.Latest.Class, "]")
	}

	// Update stored dimensions
	c.Update()
}

func (c *Client) OuterGeometry() (x, y, w, h int) {

	// Outer window dimensions (x/y relative to workspace)
	oGeom, err2 := c.Win.DecorGeometry()
	if err2 != nil {
		log.Warn(err2)
		return
	}

	// Inner window dimensions (x/y relative to outer window)
	iGeom, err1 := xwindow.RawGeometry(common.X, xproto.Drawable(c.Win.Id))
	if err1 != nil {
		log.Warn(err1)
		return
	}

	// Decoration extents (l/r/t/b relative to outer window dimensions)
	extents := c.Latest.Extents
	dx, dy, dw, dh := extents.Left, extents.Top, extents.Left+extents.Right, extents.Top+extents.Bottom

	// Calculate outer geometry (including server and client decorations)
	x, y, w, h = oGeom.X()+iGeom.X()-dx, oGeom.Y()+iGeom.Y()-dy, iGeom.Width()+dw, iGeom.Height()+dh

	return
}

func (c *Client) Update() (success bool) {
	info := GetInfo(c.Win.Id)
	if info.Class == UNKNOWN {
		return false
	}

	// Update client infos
	c.Latest = info

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
	if !motif.Decor(&c.Original.Hints) {
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
	c.Unmaximize()

	// Move window to original position
	geom := c.Original.Geometry
	err := ewmh.MoveresizeWindow(c.Win.X, c.Win.Id, geom.X(), geom.Y(), geom.Width(), geom.Height())
	if err != nil {
		log.Warn("Error when moving window [", c.Latest.Class, "]")
	}

	// Update stored dimensions
	c.Update()
}

func GetInfo(w xproto.Window) (info Info) {
	var err error

	var class string
	var name string
	var desk uint
	var states []string
	var hints *motif.Hints
	var extents *ewmh.FrameExtents
	var geometry xrect.Rect

	// Window class (internal class name of the window)
	wmClass, err := icccm.WmClassGet(common.X, w)
	if err != nil {
		log.Trace(err)
		class = UNKNOWN
	} else if wmClass != nil {
		class = wmClass.Class
	}

	// Window name (title on top of the window)
	name, err = icccm.WmNameGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		name = UNKNOWN
	}

	// Window desktop (desktop workspace where the window is visible)
	desk, err = ewmh.WmDesktopGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		desk = math.MaxUint
	}

	// Window states (visualization states of the window)
	states, err = ewmh.WmStateGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		states = []string{}
	}

	// Window hints (server decorations of the window)
	hints, err = motif.WmHintsGet(common.X, w)
	if err != nil {
		hints = &motif.Hints{}
	}

	// Window extents (server decorations of the window)
	extents, err = ewmh.FrameExtentsGet(common.X, w)
	if err != nil {
		extents = &ewmh.FrameExtents{}
	}

	// Window extents (client decorations of the window)
	ext, err := xprop.PropValNums(xprop.GetProperty(common.X, w, "_GTK_FRAME_EXTENTS"))
	if err == nil && len(ext) >= 4 {
		extents.Left -= int(ext[0])
		extents.Right -= int(ext[1])
		extents.Top -= int(ext[2])
		extents.Bottom -= int(ext[3])
	}

	// Window geometry (dimensions of the window)
	geometry, err = xwindow.New(common.X, w).DecorGeometry()
	if err != nil {
		geometry = &xrect.XRect{}
	}

	return Info{
		Class:    class,
		Name:     name,
		Desk:     desk,
		States:   states,
		Hints:    *hints,
		Extents:  *extents,
		Geometry: geometry,
	}
}

func IsMaximized(w xproto.Window) bool {
	info := GetInfo(w)
	if info.Class == UNKNOWN {
		return false
	}

	// Check maximized windows
	for _, state := range info.States {
		if strings.Contains(state, "_NET_WM_STATE_MAXIMIZED") {
			log.Info("Ignore maximized window", " [", info.Name, "]")
			return true
		}
	}

	return false
}

func IsInsideViewPort(w xproto.Window) bool {
	info := GetInfo(w)
	if info.Class == UNKNOWN {
		return false
	}

	// Viewport dimensions
	vRect := xrect.New(common.DesktopDimensions())

	// Substract viewport rectangle (r2) from window rectangle (r1)
	sRects := xrect.Subtract(info.Geometry, vRect)

	// If r1 does not overlap r2, then only one rectangle is returned which is equivalent to r1
	isOutsideViewport := false
	if len(sRects) == 1 {
		isOutsideViewport = reflect.DeepEqual(sRects[0], info.Geometry)
	}

	if isOutsideViewport {
		log.Info("Ignore window outside viewport [", info.Class, "]")
	}

	return !isOutsideViewport
}

func IsModal(info Info) bool {

	// Check model dialog windows
	for _, state := range info.States {
		if state == "_NET_WM_STATE_MODAL" {
			log.Info("Ignore modal window", " [", info.Name, "]")
			return true
		}
	}

	return false
}

func IsHidden(info Info) bool {

	// Check hidden windows
	for _, state := range info.States {
		if state == "_NET_WM_STATE_HIDDEN" {
			log.Info("Ignore hidden window", " [", info.Name, "]")
			return true
		}
	}

	return false
}

func IsFloating(info Info) bool {

	// Check floating state
	for _, state := range info.States {
		if state == "_NET_WM_STATE_ABOVE" {
			log.Info("Ignore floating window", " [", info.Name, "]")
			return true
		}
	}

	return false
}

func IsPinned(info Info) bool {

	// Check pinned windows
	if info.Desk > common.DeskCount {
		log.Info("Ignore pinned window [", info.Class, "]")
		return true
	}

	return false
}

func IsIgnored(info Info) bool {

	// Check ignored windows
	for _, s := range common.Config.WindowIgnore {
		conf_class := s[0]
		conf_name := s[1]

		reg_class := regexp.MustCompile(strings.ToLower(conf_class))
		reg_name := regexp.MustCompile(strings.ToLower(conf_name))

		// Ignore all windows with this class
		class_match := reg_class.MatchString(strings.ToLower(info.Class))

		// But allow the window with a special name
		name_match := conf_name != "" && reg_name.MatchString(strings.ToLower(info.Name))

		if class_match && !name_match {
			log.Info("Ignore window with ", strings.TrimSpace(strings.Join(s, " ")), " from config [", info.Name, "]")
			return true
		}
	}

	return false
}

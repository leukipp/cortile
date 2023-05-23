package store

import (
	"math"
	"reflect"
	"regexp"
	"strings"
	"time"

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

const (
	UNKNOWN = "<UNKNOWN>"
)

type Client struct {
	Win      *xwindow.Window `json:"-"` // X window object
	Created  time.Time       // Internal client creation time
	Latest   Info            // Latest client window information
	Original Info            // Original client window information
}

type Info struct {
	Class      string     // Client window application name
	Name       string     // Client window title name
	Desk       uint       // Client window desktop
	Types      []string   // Client window types
	States     []string   // Client window states
	Dimensions Dimensions // Client window dimensions
}

type Dimensions struct {
	Geometry xrect.Rect        // Client window geometry
	Hints    Hints             // Client window dimension hints
	Extents  ewmh.FrameExtents // Client window geometry extents
	Position bool              // Adjust position on move/resize
	Size     bool              // Adjust size on move/resize
}

type Hints struct {
	Normal icccm.NormalHints // Client window geometry hints
	Motif  motif.Hints       // Client window decoration hints
}

func CreateClient(w xproto.Window) (c *Client) {
	info := GetInfo(w)
	return &Client{
		Win:      xwindow.New(common.X, w),
		Created:  time.Now(),
		Latest:   info,
		Original: info,
	}
}

func (c *Client) Activate() {
	ewmh.ActiveWindowReq(common.X, c.Win.Id)
}

func (c *Client) MoveResize(x, y, w, h int) {
	c.UnDecorate()
	c.UnMaximize()

	// Decoration extents
	extents := c.Latest.Dimensions.Extents

	// Calculate dimensions offsets
	dx, dy, dw, dh := 0, 0, 0, 0
	if c.Latest.Dimensions.Position {
		dx, dy = extents.Left, extents.Top
	}
	if c.Latest.Dimensions.Size {
		dw, dh = extents.Left+extents.Right, extents.Top+extents.Bottom
	}

	// Move and resize window
	err := ewmh.MoveresizeWindow(common.X, c.Win.Id, x+dx, y+dy, w-dw, h-dh)
	if err != nil {
		log.Warn("Error when moving window [", c.Latest.Class, "]")
	}

	// Update stored dimensions
	c.Update()
}

func (c *Client) LimitDim(w, h int) {

	// Decoration extents
	extents := c.Latest.Dimensions.Extents
	dw, dh := extents.Left+extents.Right, extents.Top+extents.Bottom

	// Set window size limits
	icccm.WmNormalHintsSet(common.X, c.Win.Id, &icccm.NormalHints{
		Flags:     icccm.SizeHintPMinSize,
		MinWidth:  uint(w - dw),
		MinHeight: uint(h - dh),
	})
}

func (c *Client) UnDecorate() {
	if common.Config.WindowDecoration || !motif.Decor(&c.Latest.Dimensions.Hints.Motif) {
		return
	}

	// Remove window decorations
	motif.WmHintsSet(common.X, c.Win.Id, &motif.Hints{
		Flags:      motif.HintDecorations,
		Decoration: motif.DecorationNone,
	})
}

func (c *Client) UnMaximize() {

	// Unmaximize window
	for _, state := range c.Latest.States {
		if strings.Contains(state, "_NET_WM_STATE_MAXIMIZED") {
			ewmh.WmStateReq(common.X, c.Win.Id, 0, "_NET_WM_STATE_MAXIMIZED_VERT")
			ewmh.WmStateReq(common.X, c.Win.Id, 0, "_NET_WM_STATE_MAXIMIZED_HORZ")
			break
		}
	}
}

func (c *Client) Update() (success bool) {
	info := GetInfo(c.Win.Id)
	if info.Class == UNKNOWN {
		return false
	}

	// Update client info
	log.Debug("Update client info [", info.Class, "]")
	c.Latest = info

	return true
}

func (c *Client) Restore() {

	// Calculate decoration extents
	dw, dh := 0, 0
	decoration := motif.DecorationNone
	if motif.Decor(&c.Original.Dimensions.Hints.Motif) {
		decoration = motif.DecorationAll
		if !common.Config.WindowDecoration {
			extents := c.Original.Dimensions.Extents
			dw, dh = extents.Left+extents.Right, extents.Top+extents.Bottom
		}
	}

	// Restore window decorations
	motif.WmHintsSet(common.X, c.Win.Id, &motif.Hints{
		Flags:      motif.HintDecorations,
		Decoration: uint(decoration),
	})

	// Restore window size limits
	icccm.WmNormalHintsSet(common.X, c.Win.Id, &c.Original.Dimensions.Hints.Normal)

	// Move window to original position
	geom := c.Original.Dimensions.Geometry
	c.MoveResize(geom.X(), geom.Y(), geom.Width()-dw, geom.Height()-dh)
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
	extents := c.Latest.Dimensions.Extents
	dx, dy, dw, dh := extents.Left, extents.Top, extents.Left+extents.Right, extents.Top+extents.Bottom

	// Calculate outer geometry (including server and client decorations)
	x, y, w, h = oGeom.X()+iGeom.X()-dx, oGeom.Y()+iGeom.Y()-dy, iGeom.Width()+dw, iGeom.Height()+dh

	return
}

func GetInfo(w xproto.Window) (info Info) {
	var err error

	var class string
	var name string
	var desk uint
	var types []string
	var states []string
	var dimensions Dimensions

	// Window class (internal class name of the window)
	cls, err := icccm.WmClassGet(common.X, w)
	if err != nil {
		log.Trace(err)
		class = UNKNOWN
	} else if cls != nil {
		class = cls.Class
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

	// Window types (types of the window)
	types, err = ewmh.WmWindowTypeGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		types = []string{}
	}

	// Window states (states of the window)
	states, err = ewmh.WmStateGet(common.X, w)
	if err != nil {
		log.Trace(err, " [", class, "]")
		states = []string{}
	}

	// Window geometry (dimensions of the window)
	geometry, err := xwindow.New(common.X, w).DecorGeometry()
	if err != nil {
		geometry = &xrect.XRect{}
	}

	// Window normal hints (normal hints of the window)
	nhints, err := icccm.WmNormalHintsGet(common.X, w)
	if err != nil {
		nhints = &icccm.NormalHints{}
	}

	// Window motif hints (hints of the window)
	mhints, err := motif.WmHintsGet(common.X, w)
	if err != nil {
		mhints = &motif.Hints{}
	}

	// Window extents (server/client decorations of the window)
	extNet, _ := xprop.PropValNums(xprop.GetProperty(common.X, w, "_NET_FRAME_EXTENTS"))
	extGtk, _ := xprop.PropValNums(xprop.GetProperty(common.X, w, "_GTK_FRAME_EXTENTS"))

	ext := make([]uint, 4)
	for i, e := range extNet {
		ext[i] += e
	}
	for i, e := range extGtk {
		ext[i] -= e
	}

	// Window dimensions (geometry/extent information for move/resize)
	dimensions = Dimensions{
		Geometry: geometry,
		Hints: Hints{
			Normal: *nhints,
			Motif:  *mhints,
		},
		Extents: ewmh.FrameExtents{
			Left:   int(ext[0]),
			Right:  int(ext[1]),
			Top:    int(ext[2]),
			Bottom: int(ext[3]),
		},
		Position: (extNet != nil && mhints.Flags&motif.HintDecorations > 0 && mhints.Decoration > 1) || (extGtk != nil),
		Size:     (extNet != nil) || (extGtk != nil),
	}

	return Info{
		Class:      class,
		Name:       name,
		Desk:       desk,
		Types:      types,
		States:     states,
		Dimensions: dimensions,
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
			log.Info("Ignore maximized window [", info.Class, "]")
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
	sRects := xrect.Subtract(info.Dimensions.Geometry, vRect)

	// If r1 does not overlap r2, then only one rectangle is returned which is equivalent to r1
	isOutsideViewport := false
	if len(sRects) == 1 {
		isOutsideViewport = reflect.DeepEqual(sRects[0], info.Dimensions.Geometry)
	}

	if isOutsideViewport {
		log.Info("Ignore window outside viewport [", info.Class, "]")
	}

	return !isOutsideViewport
}

func IsIgnored(w xproto.Window) bool {
	info := GetInfo(w)
	if info.Class == UNKNOWN {
		return true
	}

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
			log.Info("Ignore window with ", strings.TrimSpace(strings.Join(s, " ")), " from config [", info.Class, "]")
			return true
		}
	}

	return false
}

func IsSpecial(w xproto.Window) bool {
	info := GetInfo(w)
	if info.Class == UNKNOWN {
		return true
	}

	// Check window types
	types := map[string]bool{}
	for _, typ := range []string{
		"_NET_WM_WINDOW_TYPE_DOCK",
		"_NET_WM_WINDOW_TYPE_DESKTOP",
		"_NET_WM_WINDOW_TYPE_TOOLBAR",
		"_NET_WM_WINDOW_TYPE_UTILITY",
		"_NET_WM_WINDOW_TYPE_TOOLTIP",
		"_NET_WM_WINDOW_TYPE_SPLASH",
		"_NET_WM_WINDOW_TYPE_DIALOG",
		"_NET_WM_WINDOW_TYPE_COMBO",
		"_NET_WM_WINDOW_TYPE_NOTIFICATION",
		"_NET_WM_WINDOW_TYPE_DROPDOWN_MENU",
		"_NET_WM_WINDOW_TYPE_POPUP_MENU",
		"_NET_WM_WINDOW_TYPE_MENU",
		"_NET_WM_WINDOW_TYPE_DND"} {
		types[typ] = true
	}
	for _, typ := range info.Types {
		if types[typ] {
			log.Info("Ignore window with type ", typ, " [", info.Class, "]")
			return true
		}
	}

	// Check window states
	states := map[string]bool{}
	for _, state := range []string{
		"_NET_WM_STATE_HIDDEN",
		"_NET_WM_STATE_STICKY",
		"_NET_WM_STATE_MODAL",
		"_NET_WM_STATE_ABOVE",
		"_NET_WM_STATE_BELOW",
		"_NET_WM_STATE_SKIP_PAGER",
		"_NET_WM_STATE_SKIP_TASKBAR"} {
		states[state] = true
	}
	for _, state := range info.States {
		if states[state] {
			log.Info("Ignore window with state ", state, " [", info.Class, "]")
			return true
		}
	}

	// Check pinned windows
	if info.Desk > common.DeskCount {
		log.Info("Ignore pinned window [", info.Class, "]")
		return true
	}

	return false
}

package store

import (
	"math"
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

type Client struct {
	Win      *xwindow.Window `json:"-"` // X window object
	Created  time.Time       // Internal client creation time
	Latest   *Info           // Latest client window information
	Original *Info           // Original client window information
}

type Info struct {
	Class      string     // Client window application name
	Name       string     // Client window title name
	DeskNum    uint       // Client window desktop
	ScreenNum  uint       // Client window screen
	Types      []string   // Client window types
	States     []string   // Client window states
	Dimensions Dimensions // Client window dimensions
}

type Dimensions struct {
	Geometry xrect.Rect        // Client window geometry
	Hints    Hints             // Client window dimension hints
	Extents  ewmh.FrameExtents // Client window geometry extents
	AdjPos   bool              // Adjust position on move/resize
	AdjSize  bool              // Adjust size on move/resize
}

type Hints struct {
	Normal icccm.NormalHints // Client window geometry hints
	Motif  motif.Hints       // Client window decoration hints
}

func CreateClient(w xproto.Window) *Client {
	i := GetInfo(w)
	c := &Client{
		Win:      xwindow.New(common.X, w),
		Created:  time.Now(),
		Latest:   i,
		Original: i,
	}

	// Restore window decorations
	c.Restore()

	return c
}

func (c *Client) Activate() {
	ewmh.ActiveWindowReq(common.X, c.Win.Id)
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

func (c *Client) MoveResize(x, y, w, h int) {
	c.UnDecorate()
	c.UnMaximize()

	// Decoration extents
	ext := c.Latest.Dimensions.Extents

	// Calculate dimensions offsets
	dx, dy, dw, dh := 0, 0, 0, 0
	if c.Latest.Dimensions.AdjPos {
		dx, dy = ext.Left, ext.Top
	}
	if c.Latest.Dimensions.AdjSize {
		dw, dh = ext.Left+ext.Right, ext.Top+ext.Bottom
	}

	// Move and resize window
	err := ewmh.MoveresizeWindow(common.X, c.Win.Id, x+dx, y+dy, w-dw, h-dh)
	if err != nil {
		log.Warn("Error when moving window [", c.Latest.Class, "]")
	}

	// Update stored dimensions
	c.Update()
}

func (c *Client) LimitDimensions(w, h int) {

	// Decoration extents
	ext := c.Latest.Dimensions.Extents
	dw, dh := ext.Left+ext.Right, ext.Top+ext.Bottom

	// Set window size limits
	icccm.WmNormalHintsSet(common.X, c.Win.Id, &icccm.NormalHints{
		Flags:     icccm.SizeHintPMinSize,
		MinWidth:  uint(w - dw),
		MinHeight: uint(h - dh),
	})
}

func (c *Client) Update() {
	info := GetInfo(c.Win.Id)
	if len(info.Class) == 0 {
		return
	}

	// Update client info
	log.Debug("Update client info [", info.Class, "]")
	c.Latest = info
}

func (c *Client) Restore() {
	dw, dh := 0, 0

	// Obtain decoration motif
	decoration := motif.DecorationNone
	if motif.Decor(&c.Original.Dimensions.Hints.Motif) {
		decoration = motif.DecorationAll

		// Obtain decoration extents
		if !common.Config.WindowDecoration {
			ext := c.Original.Dimensions.Extents
			dw, dh = ext.Left+ext.Right, ext.Top+ext.Bottom
		}
	}

	// Obtain dimension adjustments
	if c.Latest.Dimensions.AdjPos && c.Latest.Dimensions.AdjSize {
		c.Latest.Dimensions.AdjPos = false
		c.Latest.Dimensions.AdjSize = false
	}

	// Restore window decorations
	motif.WmHintsSet(common.X, c.Win.Id, &motif.Hints{
		Flags:      motif.HintDecorations,
		Decoration: uint(decoration),
	})

	// Restore window size limits
	icccm.WmNormalHintsSet(common.X, c.Win.Id, &c.Original.Dimensions.Hints.Normal)

	// Move window to latest position considering decoration adjustments
	geom := c.Latest.Dimensions.Geometry
	c.MoveResize(geom.X(), geom.Y(), geom.Width()-dw, geom.Height()-dh)
}

func (c *Client) OuterGeometry() (x, y, w, h int) {

	// Outer window dimensions (x/y relative to workspace)
	oGeom, err := c.Win.DecorGeometry()
	if err != nil {
		return
	}

	// Inner window dimensions (x/y relative to outer window)
	iGeom, err := xwindow.RawGeometry(common.X, xproto.Drawable(c.Win.Id))
	if err != nil {
		return
	}

	// Decoration extents (l/r/t/b relative to outer window dimensions)
	ext := c.Latest.Dimensions.Extents
	dx, dy, dw, dh := ext.Left, ext.Top, ext.Left+ext.Right, ext.Top+ext.Bottom

	// Calculate outer geometry (including server and client decorations)
	x, y, w, h = oGeom.X()+iGeom.X()-dx, oGeom.Y()+iGeom.Y()-dy, iGeom.Width()+dw, iGeom.Height()+dh

	return
}

func IsSpecial(info *Info) bool {

	// Check window types
	types := []string{
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
		"_NET_WM_WINDOW_TYPE_DND",
	}
	for _, typ := range info.Types {
		if common.IsInList(typ, types) {
			log.Info("Ignore window with type ", typ, " [", info.Class, "]")
			return true
		}
	}

	// Check window states
	states := []string{
		"_NET_WM_STATE_HIDDEN",
		"_NET_WM_STATE_STICKY",
		"_NET_WM_STATE_MODAL",
		"_NET_WM_STATE_ABOVE",
		"_NET_WM_STATE_BELOW",
		"_NET_WM_STATE_SKIP_PAGER",
		"_NET_WM_STATE_SKIP_TASKBAR",
	}
	for _, state := range info.States {
		if common.IsInList(state, states) {
			log.Info("Ignore window with state ", state, " [", info.Class, "]")
			return true
		}
	}

	// Check pinned windows
	if info.DeskNum > common.DeskCount {
		log.Info("Ignore pinned window [", info.Class, "]")
		return true
	}

	return false
}

func IsIgnored(info *Info) bool {

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

func IsMaximized(w xproto.Window) bool {
	info := GetInfo(w)

	// Check maximized windows
	for _, state := range info.States {
		if strings.Contains(state, "_NET_WM_STATE_MAXIMIZED") {
			log.Info("Ignore maximized window [", info.Class, "]")
			return true
		}
	}

	return false
}

func GetInfo(w xproto.Window) *Info {
	var err error

	var class string
	var name string
	var deskNum uint
	var screenNum uint
	var types []string
	var states []string
	var dimensions Dimensions

	// Window class (internal class name of the window)
	cls, err := icccm.WmClassGet(common.X, w)
	if err != nil {
		log.Trace("Error on request ", err)
	} else if cls != nil {
		class = cls.Class
	}

	// Window name (title on top of the window)
	name, err = icccm.WmNameGet(common.X, w)
	if err != nil {
		name = class
	}

	// Window desktop and screen (workspace where the window is located)
	deskNum, err = ewmh.WmDesktopGet(common.X, w)
	if err != nil {
		deskNum = math.MaxUint
	}
	screenNum = GetScreenNum(w)

	// Window types (types of the window)
	types, err = ewmh.WmWindowTypeGet(common.X, w)
	if err != nil {
		types = []string{}
	}

	// Window states (states of the window)
	states, err = ewmh.WmStateGet(common.X, w)
	if err != nil {
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
		AdjPos:  (extNet != nil && mhints.Flags&motif.HintDecorations > 0 && mhints.Decoration > 1) || (extGtk != nil),
		AdjSize: (extNet != nil) || (extGtk != nil),
	}

	return &Info{
		Class:      class,
		Name:       name,
		DeskNum:    deskNum,
		ScreenNum:  screenNum,
		Types:      types,
		States:     states,
		Dimensions: dimensions,
	}
}

func GetScreenNum(w xproto.Window) uint {

	// Outer window dimensions
	geom, err := xwindow.New(common.X, w).DecorGeometry()
	if err != nil {
		return 0
	}

	// Window center position
	center := &common.Pointer{
		X: int16(geom.X() + (geom.Width() / 2)),
		Y: int16(geom.Y() + (geom.Height() / 2)),
	}

	return common.ScreenNumGet(center)
}

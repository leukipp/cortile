package store

import (
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	"encoding/json"
	"io/ioutil"
	"path/filepath"

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
	Locked   bool            // Internal client move/resize lock
	Original *Info           // Original client window information
	Latest   *Info           // Latest client window information
}

type Info struct {
	Class      string     // Client window application name
	Name       string     // Client window title name
	Types      []string   // Client window types
	States     []string   // Client window states
	Location   Location   // Client window location
	Dimensions Dimensions // Client window dimensions
}

type Location struct {
	DeskNum   uint // Client workspace desktop number
	ScreenNum uint // Client workspace screen number
}

type Dimensions struct {
	Geometry   Geometry          // Client window geometry
	Hints      Hints             // Client window dimension hints
	Extents    ewmh.FrameExtents // Client window geometry extents
	AdjPos     bool              // Position adjustments on move/resize
	AdjSize    bool              // Size adjustments on move/resize
	AdjRestore bool              // Disable adjustments on restore
}

type Geometry struct {
	xrect.Rect `json:"-"` // Client window geometry functions
	X          int        // Client window geometry x position
	Y          int        // Client window geometry y position
	Width      int        // Client window geometry width dimension
	Height     int        // Client window geometry height dimension
}

type Hints struct {
	Normal icccm.NormalHints // Client window geometry hints
	Motif  motif.Hints       // Client window decoration hints
}

func CreateClient(w xproto.Window) *Client {
	i := GetInfo(w)
	c := &Client{
		Win:      xwindow.New(X, w),
		Created:  time.Now(),
		Locked:   false,
		Original: i,
		Latest:   i,
	}

	// Read client geometry from cache
	cached := c.Read()

	// Overwrite states, geometry and location
	geom := cached.Dimensions.Geometry
	geom.Rect = xrect.New(geom.X, geom.Y, geom.Width, geom.Height)

	c.Original.States = cached.States
	c.Original.Dimensions.Geometry = geom
	c.Original.Location.ScreenNum = GetScreenNum(geom.Rect)

	c.Latest.States = cached.States
	c.Latest.Dimensions.Geometry = geom
	c.Latest.Location.ScreenNum = GetScreenNum(geom.Rect)

	// Restore window position
	c.Restore(true)

	return c
}

func (c *Client) Activate() {
	ewmh.ActiveWindowReq(X, c.Win.Id)
}

func (c *Client) Lock() {
	c.Locked = true
}

func (c *Client) UnLock() {
	c.Locked = false
}

func (c *Client) UnDecorate() {
	if common.Config.WindowDecoration || !motif.Decor(&c.Latest.Dimensions.Hints.Motif) {
		return
	}

	// Remove window decorations
	mhints := c.Original.Dimensions.Hints.Motif
	mhints.Flags |= motif.HintDecorations
	mhints.Decoration = motif.DecorationNone
	motif.WmHintsSet(X, c.Win.Id, &mhints)
}

func (c *Client) UnMaximize() {

	// Unmaximize window
	for _, state := range c.Latest.States {
		if strings.HasPrefix(state, "_NET_WM_STATE_MAXIMIZED") {
			ewmh.WmStateReq(X, c.Win.Id, 0, "_NET_WM_STATE_MAXIMIZED_VERT")
			ewmh.WmStateReq(X, c.Win.Id, 0, "_NET_WM_STATE_MAXIMIZED_HORZ")
			break
		}
	}
}

func (c *Client) MoveResize(x, y, w, h int) {
	if c.Locked {
		log.Info("Reject window move/resize [", c.Latest.Class, "]")

		// Remove lock
		c.UnLock()
		return
	}

	// Remove unwanted
	c.UnDecorate()
	c.UnMaximize()

	// Calculate dimension offsets
	ext := c.Latest.Dimensions.Extents
	dx, dy, dw, dh := 0, 0, 0, 0

	if c.Latest.Dimensions.AdjPos {
		dx, dy = ext.Left, ext.Top
	}
	if c.Latest.Dimensions.AdjSize {
		dw, dh = ext.Left+ext.Right, ext.Top+ext.Bottom
	}

	// Move and resize window
	err := ewmh.MoveresizeWindow(X, c.Win.Id, x+dx, y+dy, w-dw, h-dh)
	if err != nil {
		log.Warn("Error on window move/resize [", c.Latest.Class, "]")
	}

	// Update stored dimensions
	c.Update()
}

func (c *Client) LimitDimensions(w, h int) {

	// Decoration extents
	ext := c.Latest.Dimensions.Extents
	dw, dh := ext.Left+ext.Right, ext.Top+ext.Bottom

	// Set window size limits
	nhints := c.Original.Dimensions.Hints.Normal
	nhints.Flags |= icccm.SizeHintPMinSize
	nhints.MinWidth = uint(w - dw)
	nhints.MinHeight = uint(h - dh)
	icccm.WmNormalHintsSet(X, c.Win.Id, &nhints)
}

func (c *Client) Update() {
	info := GetInfo(c.Win.Id)
	if len(info.Class) == 0 {
		return
	}
	log.Debug("Update client info [", info.Class, "]")

	// Update client info
	c.Latest = info

	// Write client cache
	c.Write()
}

func (c *Client) Write() {
	if common.Args.Cache == "0" {
		return
	}

	// Obtain cache object
	cache := c.Cache()

	// Parse client info
	data, err := json.MarshalIndent(cache.Data, "", "  ")
	if err != nil {
		log.Warn("Error parsing client info [", c.Latest.Class, "]")
		return
	}

	// Write client cache
	path := filepath.Join(cache.Folder, cache.Name)
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		log.Warn("Error writing client cache [", c.Latest.Class, "]")
		return
	}

	log.Debug("Write client cache data ", cache.Name, " [", c.Latest.Class, "]")
}

func (c *Client) Read() *Info {
	if common.Args.Cache == "0" {
		return c.Latest
	}

	// Obtain cache object
	cache := c.Cache()

	// Read client info
	path := filepath.Join(cache.Folder, cache.Name)
	data, err := ioutil.ReadFile(path)
	if os.IsNotExist(err) {
		log.Info("No client cache found [", c.Latest.Class, "]")
		return c.Latest
	}

	// Parse client info
	var info *Info
	err = json.Unmarshal([]byte(data), &info)
	if err != nil {
		log.Warn("Error reading client cache [", c.Latest.Class, "]")
		return c.Latest
	}

	log.Debug("Read client cache data ", cache.Name, " [", c.Latest.Class, "]")

	return info
}

func (c *Client) Cache() common.Cache[*Info] {

	// Create client cache folder
	folder := filepath.Join(common.Args.Cache, "clients", c.Latest.Class)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, 0700)
	}

	// Create client cache object
	hash := common.Hash(c.Latest.Class)
	cache := common.Cache[*Info]{
		Folder: folder,
		Name:   hash + ".json",
		Data:   c.Latest,
	}

	return cache
}

func (c *Client) Restore(original bool) {

	// Restore window size limits
	icccm.WmNormalHintsSet(X, c.Win.Id, &c.Original.Dimensions.Hints.Normal)

	// Restore window decorations
	motif.WmHintsSet(X, c.Win.Id, &c.Original.Dimensions.Hints.Motif)

	// Restore window states
	if common.IsInList("_NET_WM_STATE_STICKY", c.Latest.States) {
		ewmh.WmStateReq(X, c.Win.Id, 1, "_NET_WM_STATE_STICKY")
		ewmh.WmDesktopSet(X, c.Win.Id, ^uint(0))
		log.Error("Restore sticky:", c.Latest.Class)
	}

	// Disable adjustments on restore
	if c.Latest.Dimensions.AdjRestore {
		c.Latest.Dimensions.AdjPos = false
		c.Latest.Dimensions.AdjSize = false
	}

	// Move window to restore position
	geom := c.Latest.Dimensions.Geometry
	if original {
		geom = c.Original.Dimensions.Geometry
	}
	c.MoveResize(geom.X, geom.Y, geom.Width, geom.Height)
}

func (c *Client) OuterGeometry() (x, y, w, h int) {

	// Outer window dimensions (x/y relative to workspace)
	oGeom, err := c.Win.DecorGeometry()
	if err != nil {
		return
	}

	// Inner window dimensions (x/y relative to outer window)
	iGeom, err := xwindow.RawGeometry(X, xproto.Drawable(c.Win.Id))
	if err != nil {
		return
	}

	// Reset inner window positions (some wm won't return x/y relative to outer window)
	if reflect.DeepEqual(oGeom, iGeom) {
		iGeom.XSet(0)
		iGeom.YSet(0)
	}

	// Decoration extents (l/r/t/b relative to outer window dimensions)
	ext := c.Latest.Dimensions.Extents
	dx, dy, dw, dh := ext.Left, ext.Top, ext.Left+ext.Right, ext.Top+ext.Bottom

	// Calculate outer geometry (including server and client decorations)
	x, y, w, h = oGeom.X()+iGeom.X()-dx, oGeom.Y()+iGeom.Y()-dy, iGeom.Width()+dw, iGeom.Height()+dh

	return
}

func IsSpecial(info *Info) bool {

	// Check internal windows
	if info.Class == common.Build.Name {
		log.Info("Ignore internal window [", info.Class, "]")
		return true
	}

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
		if strings.HasPrefix(state, "_NET_WM_STATE_MAXIMIZED") {
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
	var types []string
	var states []string
	var location Location
	var dimensions Dimensions

	// Window class (internal class name of the window)
	cls, err := icccm.WmClassGet(X, w)
	if err != nil {
		log.Trace("Error on request ", err)
	} else if cls != nil {
		class = cls.Class
	}

	// Window name (title on top of the window)
	name, err = icccm.WmNameGet(X, w)
	if err != nil {
		name = class
	}

	// Window geometry (dimensions of the window)
	geom, err := xwindow.New(X, w).DecorGeometry()
	if err != nil {
		geom = &xrect.XRect{}
	}

	// Window desktop and screen (window workspace location)
	deskNum, err := ewmh.WmDesktopGet(X, w)
	sticky := deskNum > DeskCount
	if err != nil || sticky {
		deskNum = CurrentDesktopGet(X)
	}
	location = Location{
		DeskNum:   deskNum,
		ScreenNum: GetScreenNum(geom),
	}

	// Window types (types of the window)
	types, err = ewmh.WmWindowTypeGet(X, w)
	if err != nil {
		types = []string{}
	}

	// Window states (states of the window)
	states, err = ewmh.WmStateGet(X, w)
	if err != nil {
		states = []string{}
	}
	if sticky && !common.IsInList("_NET_WM_STATE_STICKY", states) {
		states = append(states, "_NET_WM_STATE_STICKY")
	}

	// Window normal hints (normal hints of the window)
	nhints, err := icccm.WmNormalHintsGet(X, w)
	if err != nil {
		nhints = &icccm.NormalHints{}
	}

	// Window motif hints (hints of the window)
	mhints, err := motif.WmHintsGet(X, w)
	if err != nil {
		mhints = &motif.Hints{}
	}

	// Window extents (server/client decorations of the window)
	extNet, _ := xprop.PropValNums(xprop.GetProperty(X, w, "_NET_FRAME_EXTENTS"))
	extGtk, _ := xprop.PropValNums(xprop.GetProperty(X, w, "_GTK_FRAME_EXTENTS"))

	ext := make([]uint, 4)
	for i, e := range extNet {
		ext[i] += e
	}
	for i, e := range extGtk {
		ext[i] -= e
	}

	// Window dimensions (geometry/extent information for move/resize)
	dimensions = Dimensions{
		Geometry: Geometry{
			Rect:   geom,
			X:      geom.X(),
			Y:      geom.Y(),
			Width:  geom.Width(),
			Height: geom.Height(),
		},
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
		AdjPos:     !common.IsZero(extNet) && (mhints.Decoration > 1 || nhints.WinGravity > 1) || !common.IsZero(extGtk),
		AdjSize:    !common.IsZero(extNet) || !common.IsZero(extGtk),
		AdjRestore: !common.IsZero(extGtk),
	}

	return &Info{
		Class:      class,
		Name:       name,
		Types:      types,
		States:     states,
		Location:   location,
		Dimensions: dimensions,
	}
}

func GetScreenNum(geom xrect.Rect) uint {

	// Window center position
	center := &common.Pointer{
		X: int16(geom.X() + geom.Width()/2),
		Y: int16(geom.Y() + geom.Height()/2),
	}

	return ScreenNumGet(center)
}

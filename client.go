package main

import (
	"strings"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/motif"
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/blrsn/zentile/state"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	window    *xwindow.Window
	Desk      uint // Desktop the client is currently in.
	savedProp Prop // Properties that the client had, before it was tiled.
}

type Prop struct {
	Geom       xrect.Rect
	decoration bool
}

func newClient(w xproto.Window) (c Client) {
	win := xwindow.New(state.X, w)

	desk, err := ewmh.WmDesktopGet(state.X, w)
	if err != nil {
		desk = state.CurrentDesk
	}

	savedGeom, err := win.DecorGeometry()
	if err != nil {
		log.Info(err)
	}

	c = Client{
		window: win,
		Desk:   desk,
		savedProp: Prop{
			Geom:       savedGeom,
			decoration: hasDecoration(w),
		},
	}

	return c
}

func (c Client) name() string {
	name, err := ewmh.WmNameGet(state.X, c.window.Id)
	if err != nil {
		return ""
	}

	return name
}

func (c Client) MoveResize(x, y, width, height int) {
	c.Unmaximize()

	dw, dh := c.DecorDimensions()
	err := c.window.WMMoveResize(x, y, width-dw, height-dh)

	if err != nil {
		log.Info("Error when moving ", c.name(), " ", err)
	}
}

// DecorDimensions returns the width and height occupied by window decorations
func (c Client) DecorDimensions() (width int, height int) {
	cGeom, err1 := xwindow.RawGeometry(state.X, xproto.Drawable(c.window.Id))
	pGeom, err2 := c.window.DecorGeometry()

	if err1 != nil || err2 != nil {
		return
	}

	width = pGeom.Width() - cGeom.Width()
	height = pGeom.Height() - cGeom.Height()
	return
}

func (c Client) Unmaximize() {
	ewmh.WmStateReq(state.X, c.window.Id, 0, "_NET_WM_STATE_MAXIMIZED_VERT")
	ewmh.WmStateReq(state.X, c.window.Id, 0, "_NET_WM_STATE_MAXIMIZED_HORZ")
}

func (c Client) UnDecorate() {
	motif.WmHintsSet(state.X, c.window.Id,
		&motif.Hints{
			Flags:      motif.HintDecorations,
			Decoration: motif.DecorationNone,
		})
}

func (c Client) Decorate() {
	if !c.savedProp.decoration {
		return
	}

	motif.WmHintsSet(state.X, c.window.Id,
		&motif.Hints{
			Flags:      motif.HintDecorations,
			Decoration: motif.DecorationAll,
		})
}

// Restore resizes and decorates window to pre-tiling state.
func (c Client) Restore() {
	c.Decorate()
	geom := c.savedProp.Geom
	log.Info("Restoring ", c.name(), ": ", "X: ", geom.X(), " Y: ", geom.Y())
	c.MoveResize(geom.X(), geom.Y(), geom.Width(), geom.Height())
}

//  Activate makes the client the currently active window
func (c Client) Activate() {
	ewmh.ActiveWindowReq(state.X, c.window.Id)
}

// hasDecoration returns true if the window has client decorations.
func hasDecoration(wid xproto.Window) bool {
	mh, err := motif.WmHintsGet(state.X, wid)

	if err != nil {
		return true
	}

	return motif.Decor(mh)
}

// isHidden returns true if the window has been minimized.
func isHidden(w xproto.Window) bool {
	states, _ := ewmh.WmStateGet(state.X, w)
	for _, state := range states {
		if state == "_NET_WM_STATE_HIDDEN" {
			return true
		}
	}

	return false
}

func shouldIgnore(w xproto.Window) bool {
	c, err := icccm.WmClassGet(state.X, w)
	if err != nil {
		log.Warn(err)
	}

	for _, s := range Config.WindowsToIgnore {
		if strings.EqualFold(c.Class, s) {
			return true
		}
	}

	return false
}

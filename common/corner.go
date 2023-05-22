package common

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/xrect"
)

type Corner struct {
	Name   string     // Corner name used in config
	Active bool       // Mouse pointer is in this corner
	Area   xrect.Rect // Rectangle area of the corner section
}

func CreateCorner(name string, x int, y int, w int, h int) *Corner {
	return &Corner{
		Name:   name,
		Active: false,
		Area:   xrect.New(x, y, w, h),
	}
}

func CreateCorners() []*Corner {
	xw, yw, ww, hw := ScreenDimensions()

	// Corner dimensions
	wcs, hcs := Config.EdgeCornerSize, Config.EdgeCornerSize
	wcl, hcl := Config.EdgeCenterSize, Config.EdgeCenterSize

	// Define corners and positions
	tl := CreateCorner("top_left", xw, yw, wcs, hcs)
	tc := CreateCorner("top_center", (xw+ww)/2-wcl/2, yw, wcl, hcs)
	tr := CreateCorner("top_right", xw+ww-wcs, yw, wcs, hcs)
	cr := CreateCorner("center_right", xw+ww-wcs, (yw+hw)/2-hcl/2, wcs, hcl)
	br := CreateCorner("bottom_right", xw+ww-wcs, yw+hw-hcs, wcs, hcs)
	bc := CreateCorner("bottom_center", (xw+ww)/2-wcl/2, yw+hw-hcs, wcl, hcl)
	bl := CreateCorner("bottom_left", xw, yw+hw-hcs, wcs, hcs)
	cl := CreateCorner("center_left", xw, (yw+hw)/2-hcl/2, wcs, hcl)

	return []*Corner{tl, tc, tr, cr, br, bc, bl, cl}
}

func (c *Corner) IsActive(p *xproto.QueryPointerReply) bool {

	// Check if pointer is inside rectangle
	c.Active = IsInsideRect(p, c.Area)

	return c.Active
}

func IsInsideRect(p *xproto.QueryPointerReply, r xrect.Rect) bool {
	x, y, w, h := r.Pieces()

	// Check if x and y are inside rectangle
	xInRect := int(p.RootX) >= x && int(p.RootX) <= (x+w)
	yInRect := int(p.RootY) >= y && int(p.RootY) <= (y+h)

	return xInRect && yInRect
}

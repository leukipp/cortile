package common

import "github.com/BurntSushi/xgbutil/xrect"

type Corner struct {
	Name    string     // Corner name used in config
	Enabled bool       // Corner events are enabled
	Active  bool       // Mouse pointer is in this corner
	Area    xrect.Rect // Rectangle area of the corner section
}

func CreateCorner(name string, enabled bool, x int, y int, w int, h int) (c Corner) {
	c = Corner{
		Name:    name,
		Enabled: enabled,
		Active:  false,
		Area:    xrect.New(x, y, w, h),
	}

	return c
}

func CreateCorners() []Corner {
	xw, yw, ww, hw := ScreenDimensions()

	// TODO: Load from config
	enabled := true
	wcs, hcs := 10, 10
	wcl, hcl := 100, 100

	// Define corners and positions
	tl := CreateCorner("top_left", enabled, xw, yw, wcs, hcs)
	tc := CreateCorner("top_center", enabled, (xw+ww)/2-wcl/2, yw, wcl, hcs)
	tr := CreateCorner("top_right", enabled, xw+ww-wcs, yw, wcs, hcs)
	cr := CreateCorner("center_right", enabled, xw+ww-wcs, (yw+hw)/2-hcl/2, wcs, hcl)
	br := CreateCorner("bottom_right", enabled, xw+ww-wcs, yw+hw-hcs, wcs, hcs)
	bc := CreateCorner("bottom_center", enabled, (xw+ww)/2-wcl/2, yw+hw-hcs, wcl, hcl)
	bl := CreateCorner("bottom_left", enabled, xw, yw+hw-hcs, wcs, hcs)
	cl := CreateCorner("center_left", enabled, xw, (yw+hw)/2-hcl/2, wcs, hcl)

	return []Corner{tl, tc, tr, cr, br, bc, bl, cl}
}

func (c *Corner) IsActive(p Position) bool {

	// Check if enabled and inside rectangle
	c.Active = c.Enabled && IsInsideRect(p, c.Area)

	return c.Active
}

func IsInsideRect(p Position, r xrect.Rect) bool {
	xc, yc, wc, hc := r.Pieces()

	// Check if x and y are inside rectangle
	xInRect := p.X >= xc && p.X <= (xc+wc)
	yInRect := p.Y >= yc && p.Y <= (yc+hc)

	return xInRect && yInRect
}

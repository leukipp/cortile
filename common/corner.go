package common

type Corner struct {
	Name    string // Corner name used in config
	Enabled bool   // Corner events are handled
	Active  bool   // Mouse pointer is in this corner
	Area    Area   // Area of the corner section
}

type Area struct {
	X1 uint // Rectangle top left x position
	Y1 uint // Rectangle top left y position
	X2 uint // Rectangle bottom right x position
	Y2 uint // Rectangle bottom right y position
}

func CreateCorner(name string, enabled bool, x1 int, y1 int, x2 int, y2 int) (c Corner) {
	c = Corner{
		Name:    name,
		Enabled: enabled,
		Active:  false,
		Area: Area{
			X1: uint(x1),
			Y1: uint(y1),
			X2: uint(x2),
			Y2: uint(y2),
		},
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
	tl := CreateCorner("top_left", enabled, xw, yw, xw+wcs, yw+hcs)
	tc := CreateCorner("top_center", enabled, (xw+ww)/2-wcl/2, yw, (xw+ww)/2+wcl/2, yw+hcs)
	tr := CreateCorner("top_right", enabled, xw+ww-wcs, yw, xw+ww, yw+hcs)
	cr := CreateCorner("center_right", enabled, xw+ww-wcs, (yw+hw)/2-hcl/2, xw+ww, (yw+hw)/2+hcl/2)
	br := CreateCorner("bottom_right", enabled, xw+ww-wcs, yw+hw-hcs, xw+ww, yw+hw)
	bc := CreateCorner("bottom_center", enabled, (xw+ww)/2-wcl/2, yw+hw-hcs, (xw+ww)/2+wcl/2, yw+hw)
	bl := CreateCorner("bottom_left", enabled, xw, yw+hw-hcs, xw+wcs, yw+hw)
	cl := CreateCorner("center_left", enabled, xw, (yw+hw)/2-hcl/2, xw+wcs, (yw+hw)/2+hcl/2)

	return []Corner{tl, tc, tr, cr, br, bc, bl, cl}
}

func (c *Corner) IsActive(x uint, y uint) bool {
	x1, y1, x2, y2 := c.Area.X1, c.Area.Y1, c.Area.X2, c.Area.Y2

	// Check if enabled and inside rectangle
	c.Active = c.Enabled && x >= x1 && x <= x2 && y >= y1 && y <= y2

	return c.Active
}

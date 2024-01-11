package store

import (
	"github.com/BurntSushi/xgbutil/xrect"
	"github.com/leukipp/cortile/v2/common"
)

type Corner struct {
	Name      string     // Corner name used in config
	Active    bool       // Mouse pointer is in this corner
	ScreenNum uint       // Screen number the corner is located
	Area      xrect.Rect // Rectangle area of the corner section
}

func CreateCorner(name string, screenNum uint, x int, y int, w int, h int) *Corner {
	return &Corner{
		Name:      name,
		ScreenNum: screenNum,
		Area:      xrect.New(x, y, w, h),
		Active:    false,
	}
}

func CreateCorners() []*Corner {
	corners := []*Corner{}

	for i, s := range Displays.Screens {
		xw, yw, ww, hw := s.Pieces()

		// Corner dimensions
		wcs, hcs := common.Config.EdgeCornerSize, common.Config.EdgeCornerSize
		wcl, hcl := common.Config.EdgeCenterSize, common.Config.EdgeCenterSize

		// Define corners and positions
		tl := CreateCorner("top_left", uint(i), xw, yw, wcs, hcs)
		tc := CreateCorner("top_center", uint(i), (xw+ww)/2-wcl/2, yw, wcl, hcs)
		tr := CreateCorner("top_right", uint(i), xw+ww-wcs, yw, wcs, hcs)
		cr := CreateCorner("center_right", uint(i), xw+ww-wcs, (yw+hw)/2-hcl/2, wcs, hcl)
		br := CreateCorner("bottom_right", uint(i), xw+ww-wcs, yw+hw-hcs, wcs, hcs)
		bc := CreateCorner("bottom_center", uint(i), (xw+ww)/2-wcl/2, yw+hw-hcs, wcl, hcl)
		bl := CreateCorner("bottom_left", uint(i), xw, yw+hw-hcs, wcs, hcs)
		cl := CreateCorner("center_left", uint(i), xw, (yw+hw)/2-hcl/2, wcs, hcl)

		corners = append(corners, []*Corner{tl, tc, tr, cr, br, bc, bl, cl}...)
	}

	return corners
}

func (c *Corner) IsActive(p *common.Pointer) bool {

	// Check if pointer is inside rectangle
	c.Active = common.IsInsideRect(p, c.Area)

	return c.Active
}

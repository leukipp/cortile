package store

import (
	"github.com/jezek/xgbutil/xrect"

	"github.com/leukipp/cortile/v2/common"
)

type Corner struct {
	Name      string     // Corner name used in config
	ScreenNum uint       // Screen number the corner is located
	Area      xrect.Rect // Rectangle area of the corner section
	Active    bool       // Mouse pointer is in this corner
}

func CreateCorner(name string, screenNum uint, x int, y int, w int, h int) *Corner {
	return &Corner{
		Name:      name,
		ScreenNum: screenNum,
		Area:      xrect.New(x, y, w, h),
		Active:    false,
	}
}

func CreateCorners(screens []*XHead) []*Corner {
	corners := []*Corner{}

	for i, screen := range screens {
		screenNum := uint(i)
		xw, yw, ww, hw := screen.Pieces()

		// Corner dimensions
		wcs, hcs := common.Config.EdgeCornerSize, common.Config.EdgeCornerSize
		wcl, hcl := common.Config.EdgeCenterSize, common.Config.EdgeCenterSize

		// Define corners and positions
		tl := CreateCorner("top_left", screenNum, xw, yw, wcs, hcs)
		tc := CreateCorner("top_center", screenNum, (xw+ww)/2-wcl/2, yw, wcl, hcs)
		tr := CreateCorner("top_right", screenNum, xw+ww-wcs, yw, wcs, hcs)
		cr := CreateCorner("center_right", screenNum, xw+ww-wcs, (yw+hw)/2-hcl/2, wcs, hcl)
		br := CreateCorner("bottom_right", screenNum, xw+ww-wcs, yw+hw-hcs, wcs, hcs)
		bc := CreateCorner("bottom_center", screenNum, (xw+ww)/2-wcl/2, yw+hw-hcs, wcl, hcl)
		bl := CreateCorner("bottom_left", screenNum, xw, yw+hw-hcs, wcs, hcs)
		cl := CreateCorner("center_left", screenNum, xw, (yw+hw)/2-hcl/2, wcs, hcl)

		corners = append(corners, []*Corner{tl, tc, tr, cr, br, bc, bl, cl}...)
	}

	return corners
}

func (c *Corner) IsActive(p *XPointer) bool {

	// Check if pointer is inside rectangle
	c.Active = common.IsInsideRect(p.Position, c.Area)

	return c.Active
}

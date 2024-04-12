package store

import (
	"github.com/leukipp/cortile/v2/common"

	log "github.com/sirupsen/logrus"
)

type Corner struct {
	Name      string          // Corner name used in config
	Active    bool            // Mouse pointer is in this corner
	ScreenNum uint            // Screen number the corner is located
	Geometry  common.Geometry // Geometry of the corner section
}

func CreateCorner(name string, screenNum uint, x int, y int, w int, h int) *Corner {
	return &Corner{
		Name:      name,
		Active:    false,
		ScreenNum: screenNum,
		Geometry: common.Geometry{
			X:      x,
			Y:      y,
			Width:  w,
			Height: h,
		},
	}
}

func CreateCorners(screens []XHead) []*Corner {
	corners := []*Corner{}

	for i, screen := range screens {
		screenNum := uint(i)
		x, y, w, h := screen.Geometry.Pieces()

		// Corner dimensions
		ws, hs := common.Config.EdgeCornerSize, common.Config.EdgeCornerSize
		wl, hl := common.Config.EdgeCenterSize, common.Config.EdgeCenterSize

		// Define corners and positions
		tl := CreateCorner("top_left", screenNum, x, y, ws, hs)
		tc := CreateCorner("top_center", screenNum, x+w/2-wl/2, y, wl, hs)
		tr := CreateCorner("top_right", screenNum, x+w-ws, y, ws, hs)
		cr := CreateCorner("center_right", screenNum, x+w-ws, y+h/2-hl/2, ws, hl)
		br := CreateCorner("bottom_right", screenNum, x+w-ws, y+h-hs, ws, hs)
		bc := CreateCorner("bottom_center", screenNum, x+w/2-wl/2, y+h-hs, wl, hl)
		bl := CreateCorner("bottom_left", screenNum, x, y+h-hs, ws, hs)
		cl := CreateCorner("center_left", screenNum, x, y+h/2-hl/2, ws, hl)

		corners = append(corners, []*Corner{tl, tc, tr, cr, br, bc, bl, cl}...)
	}

	return corners
}

func (c *Corner) IsActive(p *XPointer) bool {

	// Check if pointer is inside rectangle
	c.Active = common.IsInsideRect(p.Position, c.Geometry)

	return c.Active
}

func HotCorner() *Corner {

	// Update active states
	for i := range Workplace.Displays.Corners {
		hc := Workplace.Displays.Corners[i]

		wasActive := hc.Active
		isActive := hc.IsActive(Pointer)

		// Corner is hot
		if !wasActive && isActive {
			log.Debug("Corner at position ", hc.Geometry, " is hot [", hc.Name, "]")
			return hc
		}

		// Corner was hot
		if wasActive && !isActive {
			log.Debug("Corner at position ", hc.Geometry, " is cold [", hc.Name, "]")
		}
	}

	return nil
}

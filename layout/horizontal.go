package layout

import (
	"math"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type HorizontalLayout struct {
	*store.Manager        // Layout store manager
	Name           string // Layout name
}

func CreateHorizontalTopLayout(deskNum uint) *HorizontalLayout {
	manager := store.CreateManager(deskNum)
	manager.SetProportions(manager.Proportions.MasterSlave, common.Config.Proportion, 0, 1)

	return &HorizontalLayout{
		Manager: manager,
		Name:    "horizontal-top",
	}
}

func CreateHorizontalBottomLayout(deskNum uint) *HorizontalLayout {
	manager := store.CreateManager(deskNum)
	manager.SetProportions(manager.Proportions.MasterSlave, common.Config.Proportion, 1, 0)

	return &HorizontalLayout{
		Manager: manager,
		Name:    "horizontal-bottom",
	}
}

func (l *HorizontalLayout) Do() {
	log.Info("Tile ", len(l.Clients()), " windows with ", l.Name, " layout [workspace-", l.DeskNum, "]")

	dx, dy, dw, dh := common.DesktopDimensions()

	gap := common.Config.WindowGapSize

	mmax := l.Masters.Allowed
	smax := l.Slaves.Allowed

	msize := int(math.Min(float64(len(l.Masters.Clients)), float64(mmax)))
	ssize := int(math.Min(float64(len(l.Slaves.Clients)), float64(smax)))
	csize := len(l.Clients())

	my := dy
	mh := int(math.Round(float64(dh) * l.Proportions.MasterSlave[0]))
	sy := my + mh
	sh := dh - mh

	// Swap values if master is on bottom
	if l.Name == "horizontal-bottom" && csize > mmax {
		mytmp := my
		mhtmp := mh
		sytmp := sy
		shtmp := sh

		my = sytmp
		mh = shtmp
		sy = mytmp + gap
		sh = mhtmp
	}

	// Master area layout
	if msize > 0 {
		if ssize == 0 {
			mh = dh
		}

		mx := 0
		for i, c := range l.Masters.Clients {

			// Reset x position
			if i%mmax == 0 {
				mx = dx + gap
			}

			// Calculate width
			mp := l.Proportions.MasterMaster[i%msize]
			mw := int(math.Round(float64(dw-(msize+1)*gap) * mp))

			// Move and resize master
			c.MoveResize(mx, my+gap, mw, mh-2*gap)

			// Add x offset
			mx += mw + gap
		}
	}

	// Slave area layout
	if ssize > 0 {
		if msize == 0 {
			sy, sh = dy+gap, dh-gap
		}

		sx := 0
		for i, c := range l.Slaves.Clients {

			// Reset x position
			if i%smax == 0 {
				sx = dx + gap
			}

			// Calculate width
			sp := l.Proportions.SlaveSlave[i%ssize]
			sw := int(math.Round(float64(dw-(ssize+1)*gap) * sp))

			// Move and resize slave
			c.MoveResize(sx, sy, sw, sh-gap)

			// Add x offset
			sx += sw + gap
		}
	}

	common.X.Conn().Sync()
}

func (l *HorizontalLayout) UpdateProportions(c *store.Client, d *store.Directions) {
	_, _, dw, dh := common.DesktopDimensions()
	_, _, cw, ch := c.OuterGeometry()

	gap := common.Config.WindowGapSize

	mmax := l.Masters.Allowed
	smax := l.Slaves.Allowed

	msize := int(math.Min(float64(len(l.Masters.Clients)), float64(mmax)))
	ssize := int(math.Min(float64(len(l.Slaves.Clients)), float64(smax)))

	// Swap values if master is on top
	idxms := 0
	if l.Name == "horizontal-top" {
		ttmp := d.Top
		btmp := d.Bottom

		d.Top = btmp
		d.Bottom = ttmp

		idxms = 1
	}
	if l.IsMaster(c) {
		idxms ^= 1
	}

	// Calculate proportions based on window geometry
	if l.IsMaster(c) {
		py := float64(ch+2*gap) / float64(dh)
		px := float64(cw) / float64(dw-(msize+1)*gap)
		idxmm := l.Index(l.Masters.Clients, c) % mmax

		// Set master-slave proportions
		if d.Top {
			l.Manager.SetProportions(l.Proportions.MasterSlave, py, idxms, idxms^1)
		}

		// Set master-master proportions
		if d.Left {
			l.Manager.SetProportions(l.Proportions.MasterMaster, px, idxmm, idxmm-1)
		} else if d.Right {
			l.Manager.SetProportions(l.Proportions.MasterMaster, px, idxmm, idxmm+1)
		}
	} else {
		py := float64(ch+gap) / float64(dh)
		px := float64(cw) / float64(dw-(ssize+1)*gap)
		idxss := l.Index(l.Slaves.Clients, c) % smax

		// Set master-slave proportions
		if d.Bottom {
			l.Manager.SetProportions(l.Proportions.MasterSlave, py, idxms, idxms^1)
		}

		// Set slave-slave proportions
		if d.Left {
			l.Manager.SetProportions(l.Proportions.SlaveSlave, px, idxss, idxss-1)
		} else if d.Right {
			l.Manager.SetProportions(l.Proportions.SlaveSlave, px, idxss, idxss+1)
		}
	}
}

func (l *HorizontalLayout) GetManager() *store.Manager {
	return l.Manager
}

func (l *HorizontalLayout) GetName() string {
	return l.Name
}

package layout

import (
	"math"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

type HorizontalLayout struct {
	Name           string // Layout name
	*store.Manager        // Layout store manager
}

func CreateHorizontalTopLayout(loc store.Location) *HorizontalLayout {
	layout := &HorizontalLayout{
		Name:    "horizontal-top",
		Manager: store.CreateManager(loc),
	}
	layout.Reset()
	return layout
}

func CreateHorizontalBottomLayout(loc store.Location) *HorizontalLayout {
	layout := &HorizontalLayout{
		Name:    "horizontal-bottom",
		Manager: store.CreateManager(loc),
	}
	layout.Reset()
	return layout
}

func (l *HorizontalLayout) Reset() {
	mg := store.CreateManager(*l.Location)

	// Reset number of masters
	for l.Masters.MaxAllowed < mg.Masters.MaxAllowed {
		l.IncreaseMaster()
	}
	for l.Masters.MaxAllowed > mg.Masters.MaxAllowed {
		l.DecreaseMaster()
	}

	// Reset number of slaves
	for l.Slaves.MaxAllowed < mg.Slaves.MaxAllowed {
		l.IncreaseSlave()
	}
	for l.Slaves.MaxAllowed > mg.Slaves.MaxAllowed {
		l.DecreaseSlave()
	}

	// Reset layout proportions
	l.Manager.Proportions = mg.Proportions
}

func (l *HorizontalLayout) Apply() {
	clients := l.Clients(store.Stacked)

	dx, dy, dw, dh := store.DesktopDimensions(l.Location.ScreenNum)
	gap := common.Config.WindowGapSize

	mmax := l.Masters.MaxAllowed
	smax := l.Slaves.MaxAllowed

	msize := int(math.Min(float64(len(l.Masters.Stacked)), float64(mmax)))
	ssize := int(math.Min(float64(len(l.Slaves.Stacked)), float64(smax)))
	csize := len(clients)

	my := dy
	mh := int(math.Round(float64(dh) * l.Proportions.MasterSlave[2][0]))
	sy := my + mh
	sh := dh - mh

	log.Info("Tile ", csize, " windows with ", l.Name, " layout [workspace-", l.Location.DeskNum, "-", l.Location.ScreenNum, "]")

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
		minpw := common.Config.ProportionMin
		minph := common.Config.ProportionMin

		// Adjust sizes and proportions
		if ssize == 0 {
			mh = dh
			minph = 1.0
		}
		if msize == 1 {
			minpw = 1.0
		}

		mx := 0
		for i, c := range l.Masters.Stacked {

			// Reset x position
			if i%mmax == 0 {
				mx = dx + gap
			}

			// Limit minimum dimensions
			minw := int(math.Round(float64(dw-(msize+1)*gap) * minpw))
			minh := int(math.Round(float64(dh-2*gap) * minph))
			c.LimitDimensions(minw, minh)

			// Move and resize master
			mp := l.Proportions.MasterMaster[msize][i%msize]
			mw := int(math.Round(float64(dw-(msize+1)*gap) * mp))
			c.MoveResize(mx, my+gap, mw, mh-2*gap)

			// Add x offset
			mx += mw + gap
		}
	}

	// Slave area layout
	if ssize > 0 {
		minpw := common.Config.ProportionMin
		minph := common.Config.ProportionMin

		// Adjust sizes and proportions
		if msize == 0 {
			sy = dy + gap
			sh = dh - gap
			minph = 1.0
		}
		if ssize == 1 {
			minpw = 1.0
		}

		sx := 0
		for i, c := range l.Slaves.Stacked {

			// Reset x position
			if i%smax == 0 {
				sx = dx + gap
			}

			// Limit minimum dimensions
			minw := int(math.Round(float64(dw-(ssize+1)*gap) * minpw))
			minh := int(math.Round(float64(dh-2*gap) * minph))
			c.LimitDimensions(minw, minh)

			// Move and resize slave
			sp := l.Proportions.SlaveSlave[ssize][i%ssize]
			sw := int(math.Round(float64(dw-(ssize+1)*gap) * sp))
			c.MoveResize(sx, sy, sw, sh-gap)

			// Add x offset
			sx += sw + gap
		}
	}
}

func (l *HorizontalLayout) UpdateProportions(c *store.Client, d *store.Directions) {
	_, _, dw, dh := store.DesktopDimensions(l.Location.ScreenNum)
	_, _, cw, ch := c.OuterGeometry()

	gap := common.Config.WindowGapSize

	mmax := l.Masters.MaxAllowed
	smax := l.Slaves.MaxAllowed

	msize := int(math.Min(float64(len(l.Masters.Stacked)), float64(mmax)))
	ssize := int(math.Min(float64(len(l.Slaves.Stacked)), float64(smax)))

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
		idxmm := l.Index(l.Masters, c) % mmax

		// Set master-slave proportions
		if d.Top {
			l.Manager.SetProportions(l.Proportions.MasterSlave[2], py, idxms, idxms^1)
		}

		// Set master-master proportions
		if d.Left {
			l.Manager.SetProportions(l.Proportions.MasterMaster[msize], px, idxmm, idxmm-1)
		} else if d.Right {
			l.Manager.SetProportions(l.Proportions.MasterMaster[msize], px, idxmm, idxmm+1)
		}
	} else {
		py := float64(ch+gap) / float64(dh)
		px := float64(cw) / float64(dw-(ssize+1)*gap)
		idxss := l.Index(l.Slaves, c) % smax

		// Set master-slave proportions
		if d.Bottom {
			l.Manager.SetProportions(l.Proportions.MasterSlave[2], py, idxms, idxms^1)
		}

		// Set slave-slave proportions
		if d.Left {
			l.Manager.SetProportions(l.Proportions.SlaveSlave[ssize], px, idxss, idxss-1)
		} else if d.Right {
			l.Manager.SetProportions(l.Proportions.SlaveSlave[ssize], px, idxss, idxss+1)
		}
	}
}

func (l *HorizontalLayout) GetManager() *store.Manager {
	return l.Manager
}

func (l *HorizontalLayout) GetName() string {
	return l.Name
}

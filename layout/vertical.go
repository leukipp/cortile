package layout

import (
	"math"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

type VerticalLayout struct {
	*store.Manager        // Layout store manager
	Name           string // Layout name
}

func CreateVerticalLeftLayout(deskNum uint, screenNum uint) *VerticalLayout {
	manager := store.CreateManager(deskNum, screenNum)
	manager.SetProportions(manager.Proportions.MasterSlave, common.Config.Proportion, 0, 1)

	return &VerticalLayout{
		Manager: manager,
		Name:    "vertical-left",
	}
}

func CreateVerticalRightLayout(deskNum uint, screenNum uint) *VerticalLayout {
	manager := store.CreateManager(deskNum, screenNum)
	manager.SetProportions(manager.Proportions.MasterSlave, common.Config.Proportion, 1, 0)

	return &VerticalLayout{
		Manager: manager,
		Name:    "vertical-right",
	}
}

func (l *VerticalLayout) Apply() {
	clients := l.Clients(true)

	dx, dy, dw, dh := store.DesktopDimensions(l.ScreenNum)
	gap := common.Config.WindowGapSize

	mmax := l.Masters.MaxAllowed
	smax := l.Slaves.MaxAllowed

	msize := int(math.Min(float64(len(l.Masters.Items)), float64(mmax)))
	ssize := int(math.Min(float64(len(l.Slaves.Items)), float64(smax)))
	csize := len(clients)

	mx := dx
	mw := int(math.Round(float64(dw) * l.Proportions.MasterSlave[0]))
	sx := mx + mw
	sw := dw - mw

	log.Info("Tile ", csize, " windows with ", l.Name, " layout [workspace-", l.DeskNum, "-", l.ScreenNum, "]")

	// Swap values if master is on right
	if l.Name == "vertical-right" && csize > mmax {
		mxtmp := mx
		mwtmp := mw
		sxtmp := sx
		swtmp := sw

		mx = sxtmp
		mw = swtmp
		sx = mxtmp + gap
		sw = mwtmp
	}

	// Master area layout
	if msize > 0 {
		minpw := common.Config.ProportionMin
		minph := common.Config.ProportionMin

		// Adjust sizes and proportions
		if ssize == 0 {
			mw = dw
			minpw = 1.0
		}
		if msize == 1 {
			minph = 1.0
		}

		my := 0
		for i, c := range l.Masters.Items {

			// Reset y position
			if i%mmax == 0 {
				my = dy + gap
			}

			// Limit minimum dimensions
			minw := int(math.Round(float64(dw-2*gap) * minpw))
			minh := int(math.Round(float64(dh-(msize+1)*gap) * minph))
			c.LimitDimensions(minw, minh)

			// Move and resize master
			mp := l.Proportions.MasterMaster[i%msize]
			mh := int(math.Round(float64(dh-(msize+1)*gap) * mp))
			c.MoveResize(mx+gap, my, mw-2*gap, mh)

			// Add y offset
			my += mh + gap
		}
	}

	// Slave area layout
	if ssize > 0 {
		minpw := common.Config.ProportionMin
		minph := common.Config.ProportionMin

		// Adjust sizes and proportions
		if msize == 0 {
			sx = dx + gap
			sw = dw - gap
			minpw = 1.0
		}
		if ssize == 1 {
			minph = 1.0
		}

		sy := 0
		for i, c := range l.Slaves.Items {

			// Reset y position
			if i%smax == 0 {
				sy = dy + gap
			}

			// Limit minimum dimensions
			minw := int(math.Round(float64(dw-2*gap) * minpw))
			minh := int(math.Round(float64(dh-(ssize+1)*gap) * minph))
			c.LimitDimensions(minw, minh)

			// Move and resize slave
			sp := l.Proportions.SlaveSlave[i%ssize]
			sh := int(math.Round(float64(dh-(ssize+1)*gap) * sp))
			c.MoveResize(sx, sy, sw-gap, sh)

			// Add y offset
			sy += sh + gap
		}
	}
}

func (l *VerticalLayout) UpdateProportions(c *store.Client, d *store.Directions) {
	_, _, dw, dh := store.DesktopDimensions(l.ScreenNum)
	_, _, cw, ch := c.OuterGeometry()

	gap := common.Config.WindowGapSize

	mmax := l.Masters.MaxAllowed
	smax := l.Slaves.MaxAllowed

	msize := int(math.Min(float64(len(l.Masters.Items)), float64(mmax)))
	ssize := int(math.Min(float64(len(l.Slaves.Items)), float64(smax)))

	// Swap values if master is on left
	idxms := 0
	if l.Name == "vertical-left" {
		ltmp := d.Left
		rtmp := d.Right

		d.Left = rtmp
		d.Right = ltmp

		idxms = 1
	}
	if l.IsMaster(c) {
		idxms ^= 1
	}

	// Calculate proportions based on window geometry
	if l.IsMaster(c) {
		px := float64(cw+2*gap) / float64(dw)
		py := float64(ch) / float64(dh-(msize+1)*gap)
		idxmm := l.Index(l.Masters, c) % mmax

		// Set master-slave proportions
		if d.Left {
			l.Manager.SetProportions(l.Proportions.MasterSlave, px, idxms, idxms^1)
		}

		// Set master-master proportions
		if d.Top {
			l.Manager.SetProportions(l.Proportions.MasterMaster, py, idxmm, idxmm-1)
		} else if d.Bottom {
			l.Manager.SetProportions(l.Proportions.MasterMaster, py, idxmm, idxmm+1)
		}
	} else {
		px := float64(cw+gap) / float64(dw)
		py := float64(ch) / float64(dh-(ssize+1)*gap)
		idxss := l.Index(l.Slaves, c) % smax

		// Set master-slave proportions
		if d.Right {
			l.Manager.SetProportions(l.Proportions.MasterSlave, px, idxms, idxms^1)
		}

		// Set slave-slave proportions
		if d.Top {
			l.Manager.SetProportions(l.Proportions.SlaveSlave, py, idxss, idxss-1)
		} else if d.Bottom {
			l.Manager.SetProportions(l.Proportions.SlaveSlave, py, idxss, idxss+1)
		}
	}
}

func (l *VerticalLayout) GetManager() *store.Manager {
	return l.Manager
}

func (l *VerticalLayout) GetName() string {
	return l.Name
}

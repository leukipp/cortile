package layout

import (
	"math"
	"strings"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type VerticalLayout struct {
	*store.Manager         // Layout store manager
	Proportion     float64 // Master-slave proportion
	Name           string  // Layout name
}

func CreateVerticalLeftLayout(deskNum uint) *VerticalLayout {
	return &VerticalLayout{
		Manager:    store.CreateManager(deskNum),
		Proportion: common.Config.Proportion,
		Name:       "vertical-left",
	}
}

func CreateVerticalRightLayout(deskNum uint) *VerticalLayout {
	return &VerticalLayout{
		Manager:    store.CreateManager(deskNum),
		Proportion: 1.0 - common.Config.Proportion,
		Name:       "vertical-right",
	}
}

func (l *VerticalLayout) Do() {
	log.Info("Tile ", len(l.Clients()), " windows with ", l.GetName(), " layout [workspace-", l.DeskNum, "]")

	dx, dy, dw, dh := common.DesktopDimensions()
	msize := len(l.Masters)
	ssize := len(l.Slaves)
	csize := len(l.Clients())

	mx := dx
	mw := int(math.Round(float64(dw) * l.Proportion))
	sx := mx + mw
	sw := dw - mw

	mallowed := l.AllowedMasters
	sallowed := l.AllowedSlaves
	gap := common.Config.WindowGapSize

	// Master on right
	mright := strings.Contains(l.Name, "right")
	if mright && csize > mallowed {
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
		mh := (dh - (msize+1)*gap) / msize
		if ssize == 0 {
			mw = dw
		}

		for i, c := range l.Masters {
			if !common.Config.WindowDecoration {
				c.UnDecorate()
			}
			c.MoveResize(mx+gap, gap+dy+i*(mh+gap), mw-2*gap, mh)
		}
	}

	// Slave area layout
	if ssize > 0 {
		ssize = int(math.Min(float64(ssize), float64(sallowed)))
		sh := (dh - (ssize+1)*gap) / ssize
		if msize == 0 {
			sx, sw = dx, dw
		}

		for i, c := range l.Slaves {
			if !common.Config.WindowDecoration {
				c.UnDecorate()
			}
			c.MoveResize(sx, gap+dy+(i%sallowed)*(sh+gap), sw-gap, sh)
		}
	}

	common.X.Conn().Sync()
}

func (l *VerticalLayout) Undo() {
	for _, c := range append(l.Masters, l.Slaves...) {
		c.Restore()
	}
}

func (l *VerticalLayout) IncreaseProportion() {
	precision := 1.0 / common.Config.ProportionStep
	proportion := math.Round(l.Proportion*precision)/precision + common.Config.ProportionStep
	l.SetProportion(proportion)
}

func (l *VerticalLayout) DecreaseProportion() {
	precision := 1.0 / common.Config.ProportionStep
	proportion := math.Round(l.Proportion*precision)/precision - common.Config.ProportionStep
	l.SetProportion(proportion)
}

func (l *VerticalLayout) SetProportion(p float64) {
	l.Proportion = math.Min(math.Max(p, common.Config.ProportionMin), common.Config.ProportionMax)
}

func (l *VerticalLayout) GetManager() *store.Manager {
	return l.Manager
}

func (l *VerticalLayout) GetName() string {
	return l.Name
}

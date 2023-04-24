package layout

import (
	"math"
	"strings"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type HorizontalLayout struct {
	*store.Manager         // Layout store manager
	Proportion     float64 // Master-slave proportion
	Name           string  // Layout name
}

func CreateHorizontalTopLayout(deskNum uint) *HorizontalLayout {
	return &HorizontalLayout{
		Manager:    store.CreateManager(deskNum),
		Proportion: common.Config.Proportion,
		Name:       "horizontal-top",
	}
}

func CreateHorizontalBottomLayout(deskNum uint) *HorizontalLayout {
	return &HorizontalLayout{
		Manager:    store.CreateManager(deskNum),
		Proportion: 1.0 - common.Config.Proportion,
		Name:       "horizontal-bottom",
	}
}

func (l *HorizontalLayout) Do() {
	log.Info("Tile ", len(l.Clients()), " windows with ", l.GetName(), " layout [workspace-", l.DeskNum, "]")

	dx, dy, dw, dh := common.DesktopDimensions()
	msize := len(l.Masters)
	ssize := len(l.Slaves)
	csize := len(l.Clients())

	my := dy
	mh := int(math.Round(float64(dh) * l.Proportion))
	sy := my + mh
	sh := dh - mh

	mallowed := l.AllowedMasters
	sallowed := l.AllowedSlaves
	gap := common.Config.WindowGapSize

	// Master on bottom
	mbottom := strings.Contains(l.Name, "bottom")
	if mbottom && csize > mallowed {
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
		mw := (dw - (msize+1)*gap) / msize
		if ssize == 0 {
			mh = dh
		}

		for i, c := range l.Masters {
			if !common.Config.WindowDecoration {
				c.UnDecorate()
			}
			c.MoveResize(gap+dx+i*(mw+gap), my+gap, mw, mh-2*gap)
		}
	}

	// Slave area layout
	if ssize > 0 {
		ssize = int(math.Min(float64(ssize), float64(sallowed)))
		sw := (dw - (ssize+1)*gap) / ssize
		if msize == 0 {
			sy, sh = dy, dh
		}

		for i, c := range l.Slaves {
			if !common.Config.WindowDecoration {
				c.UnDecorate()
			}
			c.MoveResize(gap+dx+(i%sallowed)*(sw+gap), sy, sw, sh-gap)
		}
	}

	common.X.Conn().Sync()
}

func (l *HorizontalLayout) Undo() {
	for _, c := range append(l.Masters, l.Slaves...) {
		c.Restore()
	}
}

func (l *HorizontalLayout) IncreaseProportion() {
	precision := 1.0 / common.Config.ProportionStep
	proportion := math.Round(l.Proportion*precision)/precision + common.Config.ProportionStep
	l.SetProportion(proportion)
}

func (l *HorizontalLayout) DecreaseProportion() {
	precision := 1.0 / common.Config.ProportionStep
	proportion := math.Round(l.Proportion*precision)/precision - common.Config.ProportionStep
	l.SetProportion(proportion)
}

func (l *HorizontalLayout) SetProportion(p float64) {
	l.Proportion = math.Min(math.Max(p, common.Config.ProportionMin), common.Config.ProportionMax)
}

func (l *HorizontalLayout) GetManager() *store.Manager {
	return l.Manager
}

func (l *HorizontalLayout) GetName() string {
	return l.Name
}

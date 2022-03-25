package layout

import (
	"math"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type VerticalLayout struct {
	*store.Manager         // Layout store manager
	Proportion     float64 // Master-slave proportion
	WorkspaceNum   uint    // Active workspace index
	Type           string  // Layout name
}

func CreateVerticalLayout(workspaceNum uint) *VerticalLayout {
	return &VerticalLayout{
		Manager:      store.CreateManager(),
		Proportion:   1.0 - common.Config.Proportion, // TODO: LTR/RTL support
		WorkspaceNum: workspaceNum,
		Type:         "vertical",
	}
}

func (l *VerticalLayout) Do() {
	log.Info("Tile ", len(l.Clients()), " windows with ", l.GetType(), " layout [workspace-", l.WorkspaceNum, "]")

	dx, dy, dw, dh := common.DesktopDimensions()
	msize := len(l.Masters)
	ssize := len(l.Slaves)

	mx := dx
	mw := int(math.Round(float64(dw) * l.Proportion))
	sx := mx + mw
	sw := dw - mw
	gap := common.Config.WindowGap

	asize := len(l.Clients())
	fsize := l.AllowedMasters

	ltr := true // TODO: Load from config

	if ltr && asize > fsize {
		mxtmp := mx
		mwtmp := mw
		sxtmp := sx
		swtmp := sw

		mx = sxtmp
		mw = swtmp
		sx = mxtmp + gap
		sw = mwtmp
	}

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

	if ssize > 0 {
		sh := (dh - (ssize+1)*gap) / ssize
		if msize == 0 {
			sx, sw = dx, dw
		}

		for i, c := range l.Slaves {
			if !common.Config.WindowDecoration {
				c.UnDecorate()
			}
			c.MoveResize(sx, gap+dy+i*(sh+gap), sw-gap, sh)
		}
	}

	common.X.Conn().Sync()
}

func (l *VerticalLayout) Undo() {
	for _, c := range append(l.Masters, l.Slaves...) {
		c.Restore()
	}
}

func (l *VerticalLayout) IncrementProportion() {
	precision := 1.0 / common.Config.ProportionStep
	proportion := math.Round(l.Proportion*precision)/precision + common.Config.ProportionStep
	l.SetProportion(proportion)
}

func (l *VerticalLayout) DecrementProportion() {
	precision := 1.0 / common.Config.ProportionStep
	proportion := math.Round(l.Proportion*precision)/precision - common.Config.ProportionStep
	l.SetProportion(proportion)
}

func (l *VerticalLayout) SetProportion(p float64) {
	l.Proportion = math.Min(math.Max(p, common.Config.ProportionMin), common.Config.ProportionMax)
}

func (l *VerticalLayout) GetType() string {
	return l.Type
}

func (l *VerticalLayout) GetManager() *store.Manager {
	return l.Manager
}

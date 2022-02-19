package layout

import (
	"math"

	"github.com/leukipp/Cortile/common"
	"github.com/leukipp/Cortile/store"

	log "github.com/sirupsen/logrus"
)

type VerticalLayout struct {
	*store.Manager
	Proportion   float64
	WorkspaceNum uint
	Type         string
}

func CreateVerticalLayout(workspaceNum uint) *VerticalLayout {
	return &VerticalLayout{
		Manager:      store.CreateManager(),
		Proportion:   1.0 - common.Config.Division, // TODO: LTR/RTL support
		WorkspaceNum: workspaceNum,
		Type:         "vertical",
	}
}

func (l *VerticalLayout) Do() {
	log.Info("Tile ", len(l.All()), " windows with ", l.GetType(), " layout")

	wx, wy, ww, wh := common.WorkAreaDimensions(l.WorkspaceNum)
	msize := len(l.Masters)
	ssize := len(l.Slaves)

	mx := wx
	mw := int(math.Round(float64(ww) * l.Proportion))
	sx := mx + mw
	sw := ww - mw
	gap := common.Config.Gap

	asize := len(l.All())
	fsize := l.AllowedMasters

	// swap master-slave area for LTR/RTL support (TODO: add to config)
	swap := true

	if swap && asize > fsize {
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
		mh := (wh - (msize+1)*gap) / msize
		if ssize == 0 {
			mw = ww
		}

		for i, c := range l.Masters {
			if common.Config.HideDecor {
				c.UnDecorate()
			}
			c.MoveResize(mx+gap, gap+wy+i*(mh+gap), mw-2*gap, mh)
		}
	}

	if ssize > 0 {
		sh := (wh - (ssize+1)*gap) / ssize
		if msize == 0 {
			sx, sw = wx, ww
		}

		for i, c := range l.Slaves {
			if common.Config.HideDecor {
				c.UnDecorate()
			}
			c.MoveResize(sx, gap+wy+i*(sh+gap), sw-gap, sh)
		}
	}

	common.X.Conn().Sync()
}

func (l *VerticalLayout) Undo() {
	for _, c := range append(l.Masters, l.Slaves...) {
		c.Restore()
	}
}

func (l *VerticalLayout) NextClient() {
	l.Next().Activate()
}

func (l *VerticalLayout) PreviousClient() {
	l.Previous().Activate()
}

func (l *VerticalLayout) IncrementProportion() {
	precision := 1.0 / common.Config.Proportion
	proportion := math.Round(l.Proportion*precision)/precision + common.Config.Proportion
	l.SetProportion(proportion)
}

func (l *VerticalLayout) DecrementProportion() {
	precision := 1.0 / common.Config.Proportion
	proportion := math.Round(l.Proportion*precision)/precision - common.Config.Proportion
	l.SetProportion(proportion)
}

func (l *VerticalLayout) SetProportion(p float64) {
	l.Proportion = math.Min(math.Max(p, 0.1), 0.9)
}

func (l *VerticalLayout) GetType() string {
	return l.Type
}

func (l *VerticalLayout) GetManager() *store.Manager {
	return l.Manager
}

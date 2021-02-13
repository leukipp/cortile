package main

import (
	"github.com/blrsn/zentile/state"
	log "github.com/sirupsen/logrus"
)

type VerticalLayout struct {
	*VertHorz
}

func (l *VerticalLayout) Do() {
	log.Info("Switching to Vertical Layout")
	wx, wy, ww, wh := state.WorkAreaDimensions(l.WorkspaceNum)
	msize := len(l.masters)
	ssize := len(l.slaves)

	mx := wx
	mw := int(float64(ww) * l.Proportion)
	sx := mx + mw
	sw := ww - mw
	gap := Config.Gap

	if msize > 0 {
		mh := (wh - (msize+1)*gap) / msize
		if ssize == 0 {
			mw = ww
		}

		for i, c := range l.masters {
			if Config.HideDecor {
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

		for i, c := range l.slaves {
			if Config.HideDecor {
				c.UnDecorate()
			}
			c.MoveResize(sx, gap+wy+i*(sh+gap), sw-gap, sh)
		}
	}

	state.X.Conn().Sync()
}

type HorizontalLayout struct {
	*VertHorz
}

func (l *HorizontalLayout) Do() {
	log.Info("Switching to Horizontal Layout")
	wx, wy, ww, wh := state.WorkAreaDimensions(l.WorkspaceNum)
	msize := len(l.masters)
	ssize := len(l.slaves)

	my := wy
	mh := int(float64(wh) * l.Proportion)
	sy := my + mh
	sh := wh - mh
	gap := Config.Gap

	if msize > 0 {
		mw := (ww - (msize+1)*gap) / msize
		if ssize == 0 {
			mh = wh
		}

		for i, c := range l.masters {
			if Config.HideDecor {
				c.UnDecorate()
			}
			c.MoveResize(gap+wx+i*(mw+gap), my+gap, mw, mh-2*gap)
		}
	}

	if ssize > 0 {
		sw := (ww - (ssize+1)*gap) / ssize
		if msize == 0 {
			sy, sh = wy, wh
		}

		for i, c := range l.slaves {
			if Config.HideDecor {
				c.UnDecorate()
			}
			c.MoveResize(gap+wx+i*(sw+gap), sy, sw, sh-gap)
		}
	}

	state.X.Conn().Sync()
}

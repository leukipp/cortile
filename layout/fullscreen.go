package layout

import (
	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type FullscreenLayout struct {
	*store.Manager        // Layout store manager
	WorkspaceNum   uint   // Active workspace index
	Type           string // Layout name
}

func CreateFullscreenLayout(workspaceNum uint) *FullscreenLayout {
	return &FullscreenLayout{
		Manager:      store.CreateManager(),
		WorkspaceNum: workspaceNum,
		Type:         "fullscreen",
	}
}

func (l *FullscreenLayout) Do() {
	log.Info("Tile ", len(l.Clients()), " windows with ", l.GetType(), " layout")

	gap := common.Config.WindowGap
	for _, c := range l.Clients() {
		dx, dy, dw, dh := common.DesktopDimensions()
		c.MoveResize(dx+gap, dy+gap, dw-2*gap, dh-2*gap)
	}
}

func (l *FullscreenLayout) Undo() {
	for _, c := range append(l.Masters, l.Slaves...) {
		c.Restore()
	}
}

func (l *FullscreenLayout) NextClient() {
	l.Next().Activate()
}

func (l *FullscreenLayout) PreviousClient() {
	l.Previous().Activate()
}

func (l *FullscreenLayout) IncrementProportion() {
}

func (l *FullscreenLayout) DecrementProportion() {
}

func (l *FullscreenLayout) SetProportion(p float64) {
}

func (l *FullscreenLayout) GetType() string {
	return l.Type
}

func (l *FullscreenLayout) GetManager() *store.Manager {
	return l.Manager
}

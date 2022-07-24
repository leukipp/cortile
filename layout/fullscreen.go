package layout

import (
	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type FullscreenLayout struct {
	*store.Manager        // Layout store manager
	WorkspaceNum   uint   // Active workspace index
	Name           string // Layout name
}

func CreateFullscreenLayout(workspaceNum uint) *FullscreenLayout {
	return &FullscreenLayout{
		Manager:      store.CreateManager(),
		WorkspaceNum: workspaceNum,
		Name:         "fullscreen",
	}
}

func (l *FullscreenLayout) Do() {
	log.Info("Tile ", len(l.Clients()), " windows with ", l.GetName(), " layout [workspace-", l.WorkspaceNum, "]")

	gap := common.Config.WindowGap

	// Main area layout
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

func (l *FullscreenLayout) IncreaseProportion() {
}

func (l *FullscreenLayout) DecreaseProportion() {
}

func (l *FullscreenLayout) SetProportion(p float64) {
}

func (l *FullscreenLayout) GetName() string {
	return l.Name
}

func (l *FullscreenLayout) GetManager() *store.Manager {
	return l.Manager
}

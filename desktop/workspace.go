package desktop

import (
	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/layout"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type Workspace struct {
	Layouts         []Layout
	IsTiling        bool
	ActiveLayoutNum uint
}

func CreateWorkspaces() map[uint]*Workspace {
	workspaces := make(map[uint]*Workspace)

	for i := uint(0); i < common.DeskCount; i++ {
		ws := Workspace{
			Layouts:         CreateLayouts(i),
			IsTiling:        common.Config.StartupTiling,
			ActiveLayoutNum: 0, // TODO: add to config
		}
		workspaces[i] = &ws
	}

	return workspaces
}

func CreateLayouts(workspaceNum uint) []Layout {
	return []Layout{
		layout.CreateVerticalLayout(workspaceNum),
		layout.CreateHorizontalLayout(workspaceNum),
		layout.CreateFullscreenLayout(workspaceNum),
	}
}

func (ws *Workspace) SetLayout(layoutNum uint) {
	ws.ActiveLayoutNum = layoutNum
}

func (ws *Workspace) ActiveLayout() Layout {
	return ws.Layouts[ws.ActiveLayoutNum]
}

// Cycle through the available layouts
func (ws *Workspace) SwitchLayout() {
	ws.ActiveLayoutNum = (ws.ActiveLayoutNum + 1) % uint(len(ws.Layouts))
	ws.ActiveLayout().Do()
}

// Adds client to all the layouts in a workspace
func (ws *Workspace) AddClient(c store.Client) {
	log.Debug("Add client [", c.Class, "]")
	for _, l := range ws.Layouts {
		l.Add(c)
	}
}

// Removes client from all the layouts in a workspace
func (ws *Workspace) RemoveClient(c store.Client) {
	log.Debug("Remove client [", c.Class, "]")
	for _, l := range ws.Layouts {
		l.Remove(c)
	}
}

// Check if client is master in active layout
func (ws *Workspace) IsMaster(c store.Client) bool {
	s := ws.ActiveLayout().GetManager()
	for _, m := range s.Masters {
		if c.Win.Id == m.Win.Id {
			return true
		}
	}
	return false
}

// Tiles the active layout in a workspace
func (ws *Workspace) Tile() {
	if ws.IsTiling {
		ws.ActiveLayout().Do()
	}
}

// Untiles the active layout in a workspace.
func (ws *Workspace) UnTile() {
	ws.IsTiling = false
	ws.ActiveLayout().Undo()
}

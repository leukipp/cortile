package desktop

import (
	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/layout"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type Workspace struct {
	Layouts         []Layout // List of available layouts
	TilingEnabled   bool     // Tiling is enabled or not
	ActiveLayoutNum uint     // Active layout index
}

func CreateWorkspaces() map[uint]*Workspace {
	workspaces := make(map[uint]*Workspace)

	for i := uint(0); i < common.DeskCount; i++ {

		// Create layouts for each workspace
		layouts := CreateLayouts(i)
		ws := &Workspace{
			Layouts:       layouts,
			TilingEnabled: common.Config.TilingEnabled,
		}

		// Activate default layout
		for i, l := range layouts {
			if l.GetName() == common.Config.TilingLayout {
				ws.SetLayout(uint(i))
			}
		}

		workspaces[i] = ws
	}

	return workspaces
}

func CreateLayouts(deskNum uint) []Layout {
	return []Layout{
		layout.CreateFullscreenLayout(deskNum),
		layout.CreateVerticalLeftLayout(deskNum),
		layout.CreateVerticalRightLayout(deskNum),
		layout.CreateHorizontalTopLayout(deskNum),
		layout.CreateHorizontalBottomLayout(deskNum),
	}
}

func (ws *Workspace) SetLayout(layoutNum uint) {
	ws.ActiveLayoutNum = layoutNum
}

func (ws *Workspace) ActiveLayout() Layout {
	return ws.Layouts[ws.ActiveLayoutNum]
}

func (ws *Workspace) SwitchLayout() {
	if !ws.IsEnabled() {
		return
	}
	ws.SetLayout((ws.ActiveLayoutNum + 1) % uint(len(ws.Layouts)))
	ws.Tile()
}

func (ws *Workspace) Tile() {
	if !ws.IsEnabled() {
		return
	}
	ws.ActiveLayout().Do()
}

func (ws *Workspace) UnTile() {
	ws.ActiveLayout().Undo()
}

func (ws *Workspace) AddClient(c *store.Client) {
	if c == nil {
		return
	}

	log.Info("Add client for each layout [", c.Latest.Class, "]")

	// Add client to all layouts
	for _, l := range ws.Layouts {
		l.AddClient(c)
	}
}

func (ws *Workspace) RemoveClient(c *store.Client) {
	if c == nil {
		return
	}

	log.Info("Remove client from each layout [", c.Latest.Class, "]")

	// Remove client from all layouts
	for _, l := range ws.Layouts {
		l.RemoveClient(c)
	}
}

func (ws *Workspace) Enable(enable bool) {
	ws.TilingEnabled = enable
}

func (ws *Workspace) IsEnabled() bool {
	return ws.TilingEnabled
}

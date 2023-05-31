package desktop

import (
	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/layout"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type Workspace struct {
	Location        Location // Desktop and screen location
	Layouts         []Layout // List of available layouts
	TilingEnabled   bool     // Tiling is enabled or not
	ActiveLayoutNum uint     // Active layout index
}

func CreateWorkspaces() map[Location]*Workspace {
	workspaces := make(map[Location]*Workspace)

	for deskNum := uint(0); deskNum < common.DeskCount; deskNum++ {
		for screenNum := uint(0); screenNum < common.ScreenCount; screenNum++ {
			location := Location{DeskNum: deskNum, ScreenNum: screenNum}

			// Create layouts for each desktop and screen
			layouts := CreateLayouts(location)
			ws := &Workspace{
				Location:        location,
				Layouts:         layouts,
				TilingEnabled:   common.Config.TilingEnabled,
				ActiveLayoutNum: 0,
			}

			// Activate default layout
			for i, l := range layouts {
				if l.GetName() == common.Config.TilingLayout {
					ws.SetLayout(uint(i))
				}
			}

			// Map location to workspace
			workspaces[location] = ws
		}
	}

	return workspaces
}

func CreateLayouts(l Location) []Layout {
	return []Layout{
		layout.CreateFullscreenLayout(l.DeskNum, l.ScreenNum),
		layout.CreateVerticalLeftLayout(l.DeskNum, l.ScreenNum),
		layout.CreateVerticalRightLayout(l.DeskNum, l.ScreenNum),
		layout.CreateHorizontalTopLayout(l.DeskNum, l.ScreenNum),
		layout.CreateHorizontalBottomLayout(l.DeskNum, l.ScreenNum),
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
	log.Info("Add client for each layout [", c.Latest.Class, "]")

	// Add client to all layouts
	for _, l := range ws.Layouts {
		l.AddClient(c)
	}
}

func (ws *Workspace) RemoveClient(c *store.Client) {
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

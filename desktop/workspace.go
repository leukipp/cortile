package desktop

import (
	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/layout"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

type Workspace struct {
	Location        store.Location // Desktop and screen location
	Layouts         []Layout       // List of available layouts
	TilingEnabled   bool           // Tiling is enabled or not
	ActiveLayoutNum uint           // Active layout index
}

func CreateWorkspaces() map[store.Location]*Workspace {
	workspaces := make(map[store.Location]*Workspace)

	for deskNum := uint(0); deskNum < store.DeskCount; deskNum++ {
		for screenNum := uint(0); screenNum < store.ScreenCount; screenNum++ {
			location := store.Location{DeskNum: deskNum, ScreenNum: screenNum}

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

func CreateLayouts(l store.Location) []Layout {
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

func (ws *Workspace) CycleLayout(step int) {
	if ws.Disabled() {
		return
	}

	// Calculate cycle direction
	i := (int(ws.ActiveLayoutNum) + step) % len(ws.Layouts)
	if i < 0 {
		i = len(ws.Layouts) - 1
	}

	ws.SetLayout(uint(i))
	ws.Tile()
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

func (ws *Workspace) Tile() {
	if ws.Disabled() {
		return
	}

	// Apply active layout
	ws.ActiveLayout().Apply()
}

func (ws *Workspace) Restore(flag uint8) {
	mg := ws.ActiveLayout().GetManager()
	clients := mg.Clients(store.Stacked)

	log.Info("Untile ", len(clients), " windows [workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	// Restore client dimensions
	for _, c := range clients {
		if c == nil {
			continue
		}
		c.Restore(flag)
	}
}

func (ws *Workspace) Enable() {
	ws.TilingEnabled = true
}

func (ws *Workspace) Disable() {
	ws.TilingEnabled = false
}

func (ws *Workspace) Enabled() bool {
	if ws == nil {
		return false
	}
	return ws.TilingEnabled
}

func (ws *Workspace) Disabled() bool {
	if ws == nil {
		return false
	}
	return !ws.TilingEnabled
}

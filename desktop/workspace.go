package desktop

import (
	"fmt"
	"math"
	"os"

	"encoding/json"
	"path/filepath"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/layout"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

type Workspace struct {
	Name     string         // Workspace location name
	Location store.Location // Desktop and screen location
	Layouts  []Layout       // List of available layouts
	Layout   uint           // Active layout index
	Tiling   bool           // Tiling is enabled
}

func CreateWorkspaces() map[store.Location]*Workspace {
	workspaces := make(map[store.Location]*Workspace)

	for desktop := uint(0); desktop < store.Workplace.DesktopCount; desktop++ {
		for screen := uint(0); screen < store.Workplace.ScreenCount; screen++ {
			location := store.Location{Desktop: desktop, Screen: screen}

			// Create layouts for each desktop and screen
			ws := &Workspace{
				Name:     fmt.Sprintf("workspace-%d-%d", location.Desktop, location.Screen),
				Location: location,
				Layouts:  CreateLayouts(location),
				Layout:   0,
				Tiling:   common.Config.TilingEnabled,
			}

			// Set default layout
			for i, l := range ws.Layouts {
				if l.GetName() == common.Config.TilingLayout {
					ws.SetLayout(uint(i))
				}
			}

			// Read workspace from cache
			cached := ws.Read()

			// Overwrite default layout, proportions, decoration and tiling state
			ws.SetLayout(cached.Layout)
			for _, l := range ws.Layouts {
				for _, cl := range cached.Layouts {
					if l.GetName() == cl.GetName() {
						mg, cmg := l.GetManager(), cl.GetManager()
						mg.Masters.Maximum = int(math.Min(float64(cmg.Masters.Maximum), float64(common.Config.WindowMastersMax)))
						mg.Slaves.Maximum = int(math.Min(float64(cmg.Slaves.Maximum), float64(common.Config.WindowSlavesMax)))
						mg.Proportions = cmg.Proportions
						mg.Decoration = cmg.Decoration
					}
				}
			}
			ws.Tiling = cached.Tiling

			// Map location to workspace
			workspaces[location] = ws
		}
	}

	return workspaces
}

func CreateLayouts(loc store.Location) []Layout {
	return []Layout{
		layout.CreateVerticalLeftLayout(loc),
		layout.CreateVerticalRightLayout(loc),
		layout.CreateHorizontalTopLayout(loc),
		layout.CreateHorizontalBottomLayout(loc),
		layout.CreateMaximizedLayout(loc),
		layout.CreateFullscreenLayout(loc),
	}
}

func (ws *Workspace) EnableTiling() {
	ws.Tiling = true
}

func (ws *Workspace) DisableTiling() {
	ws.Tiling = false
}

func (ws *Workspace) TilingEnabled() bool {
	if ws == nil {
		return false
	}
	return ws.Tiling
}

func (ws *Workspace) TilingDisabled() bool {
	if ws == nil {
		return true
	}
	return !ws.Tiling
}

func (ws *Workspace) ActiveLayout() Layout {
	return ws.Layouts[ws.Layout]
}

func (ws *Workspace) SetLayout(layout uint) {
	ws.Layout = layout
}

func (ws *Workspace) ResetLayouts() {

	// Reset layouts
	for _, l := range ws.Layouts {

		// Reset client decorations
		mg := l.GetManager()
		mg.Decoration = common.Config.WindowDecoration

		// Reset layout proportions
		l.Reset()
	}
}

func (ws *Workspace) CycleLayout(step int) {

	// Calculate cycle direction
	i := (int(ws.Layout) + step) % len(ws.Layouts)
	if i < 0 {
		i = len(ws.Layouts) + step
	}

	ws.SetLayout(uint(i))
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
	if ws.TilingDisabled() {
		return
	}
	mg := ws.ActiveLayout().GetManager()
	clients := mg.Clients(store.Stacked)

	// Set client decorations
	for _, c := range clients {
		if c == nil {
			continue
		}
		if mg.DecorationEnabled() {
			if c.Decorate() {
				c.Update()
			}
		} else {
			if c.UnDecorate() {
				c.Update()
			}
		}
	}

	// Apply active layout
	ws.ActiveLayout().Apply()
}

func (ws *Workspace) Restore(flag uint8) {
	mg := ws.ActiveLayout().GetManager()
	clients := mg.Clients(store.Stacked)

	log.Info("Untile ", len(clients), " windows [", ws.Name, "]")

	// Restore client dimensions
	for _, c := range clients {
		if c == nil {
			continue
		}
		c.Restore(flag)
	}
}

func (ws *Workspace) Write() {
	if common.CacheDisabled() {
		return
	}

	// Obtain cache object
	cache := ws.Cache()

	// Parse workspace cache
	data, err := json.MarshalIndent(cache.Data, "", "  ")
	if err != nil {
		log.Warn("Error parsing workspace cache [", ws.Name, "]")
		return
	}

	// Write workspace cache
	path := filepath.Join(cache.Folder, cache.Name)
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		log.Warn("Error writing workspace cache [", ws.Name, "]")
		return
	}

	log.Trace("Write workspace cache data ", cache.Name, " [", ws.Name, "]")
}

func (ws *Workspace) Read() *Workspace {
	if common.CacheDisabled() {
		return ws
	}

	// Obtain cache object
	cache := ws.Cache()

	// Read workspace cache
	path := filepath.Join(cache.Folder, cache.Name)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		log.Info("No workspace cache found [", ws.Name, "]")
		return ws
	}

	// Parse workspace cache
	cached := &Workspace{Layouts: CreateLayouts(ws.Location)}
	err = json.Unmarshal([]byte(data), &cached)
	if err != nil {
		log.Warn("Error reading workspace cache [", ws.Name, "]")
		return ws
	}

	log.Debug("Read workspace cache data ", cache.Name, " [", ws.Name, "]")

	return cached
}

func (ws *Workspace) Cache() common.Cache[*Workspace] {
	name := fmt.Sprintf("workspace-%d", ws.Location.Desktop)
	hash := fmt.Sprintf("%s-%d", name, ws.Location.Screen)

	// Create workspace cache folder
	folder := filepath.Join(common.Args.Cache, "workplaces", store.Workplace.Displays.Name, "workspaces", name)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, 0755)
	}

	// Create workspace cache object
	cache := common.Cache[*Workspace]{
		Folder: folder,
		Name:   common.HashString(hash) + ".json",
		Data:   ws,
	}

	return cache
}

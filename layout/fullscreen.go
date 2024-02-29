package layout

import (
	"math"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

type FullscreenLayout struct {
	Name           string // Layout name
	*store.Manager        // Layout store manager
}

func CreateFullscreenLayout(loc store.Location) *FullscreenLayout {
	layout := &FullscreenLayout{
		Name:    "fullscreen",
		Manager: store.CreateManager(loc),
	}
	layout.Reset()
	return layout
}

func (l *FullscreenLayout) Reset() {
	mg := store.CreateManager(*l.Location)

	// Reset layout proportions
	l.Manager.Proportions = mg.Proportions
}

func (l *FullscreenLayout) Apply() {
	clients := l.Clients(store.Stacked)

	dx, dy, dw, dh := store.DesktopDimensions(l.Location.ScreenNum)
	gap := common.Config.WindowGapSize

	csize := len(clients)

	log.Info("Tile ", csize, " windows with ", l.Name, " layout [workspace-", l.Location.DeskNum, "-", l.Location.ScreenNum, "]")

	// Main area layout
	for _, c := range clients {

		// Limit minimum dimensions
		minw := int(math.Round(float64(dw - 2*gap)))
		minh := int(math.Round(float64(dh - 2*gap)))
		c.LimitDimensions(minw, minh)

		// Move and resize client
		c.MoveResize(dx+gap, dy+gap, dw-2*gap, dh-2*gap)
	}
}

func (l *FullscreenLayout) UpdateProportions(c *store.Client, d *store.Directions) {
	l.Reset()
}

func (l *FullscreenLayout) GetManager() *store.Manager {
	return l.Manager
}

func (l *FullscreenLayout) GetName() string {
	return l.Name
}

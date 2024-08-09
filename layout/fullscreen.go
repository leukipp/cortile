package layout

import (
	"math"

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
	clients := l.Ordered(&store.Clients{Stacked: l.Clients(store.Stacked)})

	_, _, dw, dh := store.ScreenGeometry(l.Location.ScreenNum).Pieces()

	csize := len(clients)

	log.Info("Tile ", csize, " windows with ", l.Name, " layout [workspace-", l.Location.DeskNum, "-", l.Location.ScreenNum, "]")

	// Main area layout
	for _, c := range clients {

		// Limit minimum dimensions
		minw := int(math.Round(float64(dw)))
		minh := int(math.Round(float64(dh)))
		c.LimitDimensions(minw, minh)

		// Make window fullscreen
		c.Fullscreen()
		c.Update()
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

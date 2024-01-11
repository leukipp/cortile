package layout

import (
	"math"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

type FullscreenLayout struct {
	*store.Manager        // Layout store manager
	Name           string // Layout name
}

func CreateFullscreenLayout(deskNum uint, screenNum uint) *FullscreenLayout {
	return &FullscreenLayout{
		Manager: store.CreateManager(deskNum, screenNum),
		Name:    "fullscreen",
	}
}

func (l *FullscreenLayout) Apply() {
	clients := l.Clients(true)

	dx, dy, dw, dh := store.DesktopDimensions(l.ScreenNum)
	gap := common.Config.WindowGapSize

	csize := len(clients)

	log.Info("Tile ", csize, " windows with ", l.Name, " layout [workspace-", l.DeskNum, "-", l.ScreenNum, "]")

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
	l.Proportions.MasterSlave = []float64{1.0}
}

func (l *FullscreenLayout) GetManager() *store.Manager {
	return l.Manager
}

func (l *FullscreenLayout) GetName() string {
	return l.Name
}

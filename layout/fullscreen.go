package layout

import (
	"math"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

type FullscreenLayout struct {
	*store.Manager        // Layout store manager
	Name           string // Layout name
}

func CreateFullscreenLayout(deskNum uint) *FullscreenLayout {
	return &FullscreenLayout{
		Manager: store.CreateManager(deskNum),
		Name:    "fullscreen",
	}
}

func (l *FullscreenLayout) Do() {
	clients := l.Clients(true)

	dx, dy, dw, dh := common.DesktopDimensions()
	gap := common.Config.WindowGapSize

	csize := len(clients)

	log.Info("Tile ", csize, " windows with ", l.Name, " layout [workspace-", l.DeskNum, "]")

	// Main area layout
	for _, c := range clients {

		// Limit minimum dimensions
		minw := int(math.Round(float64(dw - 2*gap)))
		minh := int(math.Round(float64(dh - 2*gap)))
		c.LimitDim(minw, minh)

		// Move and resize client
		c.MoveResize(dx+gap, dy+gap, dw-2*gap, dh-2*gap)
	}

	common.X.Conn().Sync()
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

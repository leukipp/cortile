package store

import (
	"math"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/leukipp/cortile/common"

	log "github.com/sirupsen/logrus"
)

type Manager struct {
	DeskNum     uint         // Index of managed desktop
	ScreenNum   uint         // Index of managed screen
	Proportions *Proportions // Layout proportions of window clients
	Masters     *Windows     // List of master window clients
	Slaves      *Windows     // List of slave window clients
}

type Directions struct {
	Top    bool // Indicates proportion changes on the top
	Right  bool // Indicates proportion changes on the right
	Bottom bool // Indicates proportion changes on the bottom
	Left   bool // Indicates proportion changes on the left
}

type Proportions struct {
	MasterSlave  []float64 // Master-slave proportions
	MasterMaster []float64 // Master-master proportions
	SlaveSlave   []float64 // Slave-slave proportions
}

type Windows struct {
	Clients    []*Client // List of stored window clients
	MaxAllowed int       // Number of maximal allowed clients
}

func CreateManager(deskNum uint, screenNum uint) *Manager {
	return &Manager{
		DeskNum:   deskNum,
		ScreenNum: screenNum,
		Proportions: &Proportions{
			MasterSlave:  calcProportions(2),
			MasterMaster: calcProportions(common.Config.WindowMastersMax),
			SlaveSlave:   calcProportions(common.Config.WindowSlavesMax),
		},
		Masters: &Windows{
			Clients:    make([]*Client, 0),
			MaxAllowed: 1,
		},
		Slaves: &Windows{
			Clients:    make([]*Client, 0),
			MaxAllowed: common.Config.WindowSlavesMax,
		},
	}
}

func (mg *Manager) Undo() {
	clients := mg.Clients(true)

	log.Info("Untile ", len(clients), " windows [workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	for _, c := range clients {
		c.Restore()
	}
}

func (mg *Manager) AddClient(c *Client) {
	log.Debug("Add client for manager [", c.Latest.Class, ", workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	// Fill up master area then slave area
	if len(mg.Masters.Clients) < mg.Masters.MaxAllowed {
		mg.updateMasters(addClient(mg.Masters.Clients, c))
	} else {
		mg.updateSlaves(addClient(mg.Slaves.Clients, c))
	}
}

func (mg *Manager) RemoveClient(c *Client) {
	log.Debug("Remove client from manager [", c.Latest.Class, ", workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	// Remove master window
	mi := mg.Index(mg.Masters, c)
	if mi >= 0 {
		if len(mg.Slaves.Clients) > 0 {
			mg.SwapClient(mg.Masters.Clients[mi], mg.Slaves.Clients[0])
			mg.updateSlaves(mg.Slaves.Clients[1:])
		} else {
			mg.updateMasters(removeClient(mg.Masters.Clients, mi))
		}
		return
	}

	// Remove slave window
	si := mg.Index(mg.Slaves, c)
	if si >= 0 {
		mg.updateSlaves(removeClient(mg.Slaves.Clients, si))
		return
	}
}

func (mg *Manager) MakeMaster(c *Client) {
	log.Info("Make window master [", c.Latest.Class, ", workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	// Swap window with first master
	if len(mg.Masters.Clients) > 0 {
		mg.SwapClient(c, mg.Masters.Clients[0])
	}
}

func (mg *Manager) SwapClient(c1 *Client, c2 *Client) {
	log.Info("Swap clients [", c1.Latest.Class, "-", c2.Latest.Class, ", workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	mIndex1 := mg.Index(mg.Masters, c1)
	sIndex1 := mg.Index(mg.Slaves, c1)

	mIndex2 := mg.Index(mg.Masters, c2)
	sIndex2 := mg.Index(mg.Slaves, c2)

	// Swap master with master
	if mIndex1 >= 0 && mIndex2 >= 0 {
		mg.Masters.Clients[mIndex2], mg.Masters.Clients[mIndex1] = mg.Masters.Clients[mIndex1], mg.Masters.Clients[mIndex2]
		return
	}

	// Swap master with slave
	if mIndex1 >= 0 && sIndex2 >= 0 {
		mg.Slaves.Clients[sIndex2], mg.Masters.Clients[mIndex1] = mg.Masters.Clients[mIndex1], mg.Slaves.Clients[sIndex2]
		return
	}

	// Swap slave with master
	if sIndex1 >= 0 && mIndex2 >= 0 {
		mg.Masters.Clients[mIndex2], mg.Slaves.Clients[sIndex1] = mg.Slaves.Clients[sIndex1], mg.Masters.Clients[mIndex2]
		return
	}

	// Swap slave with slave
	if sIndex1 >= 0 && sIndex2 >= 0 {
		mg.Slaves.Clients[sIndex2], mg.Slaves.Clients[sIndex1] = mg.Slaves.Clients[sIndex1], mg.Slaves.Clients[sIndex2]
		return
	}
}

func (mg *Manager) NextClient() {
	clients := mg.Clients(true)
	last := len(clients) - 1

	// Get next window
	for i, c := range clients {
		if c.Win.Id == common.ActiveWindow {
			next := i + 1
			if next > last {
				next = 0
			}
			clients[next].Activate()
			break
		}
	}
}

func (mg *Manager) PreviousClient() {
	clients := mg.Clients(true)
	last := len(clients) - 1

	// Get previous window
	for i, c := range clients {
		if c.Win.Id == common.ActiveWindow {
			prev := i - 1
			if prev < 0 {
				prev = last
			}
			clients[prev].Activate()
			break
		}
	}
}

func (mg *Manager) IncreaseMaster() {

	// Increase master area
	if len(mg.Slaves.Clients) > 1 && mg.Masters.MaxAllowed < common.Config.WindowMastersMax {
		mg.Masters.MaxAllowed += 1
		mg.updateMasters(append(mg.Masters.Clients, mg.Slaves.Clients[0]))
		mg.updateSlaves(mg.Slaves.Clients[1:])
	}

	log.Info("Increase masters to ", mg.Masters.MaxAllowed)
}

func (mg *Manager) DecreaseMaster() {

	// Decrease master area
	if len(mg.Masters.Clients) > 0 {
		mg.Masters.MaxAllowed -= 1
		mg.updateSlaves(append([]*Client{mg.Masters.Clients[len(mg.Masters.Clients)-1]}, mg.Slaves.Clients...))
		mg.updateMasters(mg.Masters.Clients[:len(mg.Masters.Clients)-1])
	}

	log.Info("Decrease masters to ", mg.Masters.MaxAllowed)
}

func (mg *Manager) IncreaseSlave() {

	// Increase slave area
	if mg.Slaves.MaxAllowed < common.Config.WindowSlavesMax {
		mg.Slaves.MaxAllowed += 1
		mg.updateSlaves(mg.Slaves.Clients)
	}

	log.Info("Increase slaves to ", mg.Slaves.MaxAllowed)
}

func (mg *Manager) DecreaseSlave() {

	// Decrease slave area
	if mg.Slaves.MaxAllowed > 1 {
		mg.Slaves.MaxAllowed -= 1
		mg.updateSlaves(mg.Slaves.Clients)
	}

	log.Info("Decrease slaves to ", mg.Slaves.MaxAllowed)
}

func (mg *Manager) IncreaseProportion() {
	precision := 1.0 / common.Config.ProportionStep

	// Increase root proportion
	proportion := math.Round(mg.Proportions.MasterSlave[0]*precision)/precision + common.Config.ProportionStep
	mg.SetProportions(mg.Proportions.MasterSlave, proportion, 0, 1)
}

func (mg *Manager) DecreaseProportion() {
	precision := 1.0 / common.Config.ProportionStep

	// Decrease root proportion
	proportion := math.Round(mg.Proportions.MasterSlave[0]*precision)/precision - common.Config.ProportionStep
	mg.SetProportions(mg.Proportions.MasterSlave, proportion, 0, 1)
}

func (mg *Manager) SetProportions(ps []float64, pi float64, i int, j int) bool {

	// Ignore changes on border sides
	if i == j || j < 0 || j >= len(ps) {
		return false
	}

	// Clamp target proportion
	pic := math.Min(math.Max(pi, common.Config.ProportionMin), 1.0-common.Config.ProportionMin)
	if pi != pic {
		return false
	}

	// Clamp neighbor proportion
	pj := ps[j] + (ps[i] - pi)
	pjc := math.Min(math.Max(pj, common.Config.ProportionMin), 1.0-common.Config.ProportionMin)
	if pj != pjc {
		return false
	}

	// Update proportions
	ps[i] = pi
	ps[j] = pj

	return true
}

func (mg *Manager) IsMaster(c *Client) bool {

	// Check if window is master
	return mg.Index(mg.Masters, c) >= 0
}

func (mg *Manager) IsSlave(c *Client) bool {

	// Check if window is slave
	return mg.Index(mg.Slaves, c) >= 0
}

func (mg *Manager) Index(windows *Windows, c *Client) int {

	// Traverse client list
	for i, m := range windows.Clients {
		if m.Win.Id == c.Win.Id {
			return i
		}
	}

	return -1
}

func (mg *Manager) Ordered(windows *Windows) []*Client {
	ordered := []*Client{}

	// Create ordered client list
	stacking, _ := ewmh.ClientListStackingGet(common.X)
	for _, w := range stacking {
		for _, c := range windows.Clients {
			if w == c.Win.Id {
				ordered = append(ordered, c)
				break
			}
		}
	}

	return ordered
}

func (mg *Manager) Visible(windows *Windows) []*Client {
	visible := make([]*Client, int(math.Min(float64(len(windows.Clients)), float64(windows.MaxAllowed))))

	// Create visible client list
	for _, c := range mg.Ordered(windows) {
		visible[mg.Index(windows, c)%windows.MaxAllowed] = c
	}

	return visible
}

func (mg *Manager) Clients(all bool) []*Client {
	if all {
		return append(mg.Masters.Clients, mg.Slaves.Clients...)
	}
	return append(mg.Visible(mg.Masters), mg.Visible(mg.Slaves)...)
}

func (mg *Manager) updateMasters(cs []*Client) {
	mg.Masters.Clients = cs
	mg.Proportions.MasterMaster = calcProportions(int(math.Min(float64(len(mg.Masters.Clients)), float64(mg.Masters.MaxAllowed))))
}

func (mg *Manager) updateSlaves(cs []*Client) {
	mg.Slaves.Clients = cs
	mg.Proportions.SlaveSlave = calcProportions(int(math.Min(float64(len(mg.Slaves.Clients)), float64(mg.Slaves.MaxAllowed))))
}

func addClient(cs []*Client, c *Client) []*Client {
	return append([]*Client{c}, cs...)
}

func removeClient(cs []*Client, i int) []*Client {
	return append(cs[:i], cs[i+1:]...)
}

func calcProportions(n int) []float64 {
	p := []float64{}
	for i := 0; i < n; i++ {
		p = append(p, 1.0/float64(n))
	}
	return p
}

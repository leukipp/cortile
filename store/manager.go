package store

import (
	"math"

	"github.com/leukipp/cortile/v2/common"

	log "github.com/sirupsen/logrus"
)

type Manager struct {
	DeskNum     uint         // Index of managed desktop
	ScreenNum   uint         // Index of managed screen
	Proportions *Proportions // Layout proportions of window clients
	Masters     *Clients     // List of master window clients
	Slaves      *Clients     // List of slave window clients
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

type Clients struct {
	Items      []*Client // List of stored window clients
	MaxAllowed int       // Currently maximum allowed clients
}

const (
	Stacked uint8 = 1 // Flag for stacked (all) clients
	Visible uint8 = 2 // Flag for visible (top) clients
)

func CreateManager(deskNum uint, screenNum uint) *Manager {
	return &Manager{
		DeskNum:   deskNum,
		ScreenNum: screenNum,
		Proportions: &Proportions{
			MasterSlave:  calcProportions(2),
			MasterMaster: calcProportions(common.Config.WindowMastersMax),
			SlaveSlave:   calcProportions(common.Config.WindowSlavesMax),
		},
		Masters: &Clients{
			Items:      make([]*Client, 0),
			MaxAllowed: int(math.Min(float64(common.Config.WindowMastersMax), 1)),
		},
		Slaves: &Clients{
			Items:      make([]*Client, 0),
			MaxAllowed: int(math.Max(float64(common.Config.WindowSlavesMax), 1)),
		},
	}
}

func (mg *Manager) AddClient(c *Client) {
	if mg.IsMaster(c) || mg.IsSlave(c) {
		return
	}

	log.Debug("Add client for manager [", c.Latest.Class, ", workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	// Fill up master area then slave area
	if len(mg.Masters.Items) < mg.Masters.MaxAllowed {
		mg.updateMasters(addClient(mg.Masters.Items, c))
	} else {
		mg.updateSlaves(addClient(mg.Slaves.Items, c))
	}
}

func (mg *Manager) RemoveClient(c *Client) {
	log.Debug("Remove client from manager [", c.Latest.Class, ", workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	// Remove master window
	mi := mg.Index(mg.Masters, c)
	if mi >= 0 {
		if len(mg.Slaves.Items) > 0 {
			mg.SwapClient(mg.Masters.Items[mi], mg.Slaves.Items[0])
			mg.updateSlaves(mg.Slaves.Items[1:])
		} else {
			mg.updateMasters(removeClient(mg.Masters.Items, mi))
		}
	}

	// Remove slave window
	si := mg.Index(mg.Slaves, c)
	if si >= 0 {
		mg.updateSlaves(removeClient(mg.Slaves.Items, si))
	}
}

func (mg *Manager) MakeMaster(c *Client) {
	log.Info("Make window master [", c.Latest.Class, ", workspace-", mg.DeskNum, "-", mg.ScreenNum, "]")

	// Swap window with first master
	if len(mg.Masters.Items) > 0 {
		mg.SwapClient(c, mg.Masters.Items[0])
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
		mg.Masters.Items[mIndex2], mg.Masters.Items[mIndex1] = mg.Masters.Items[mIndex1], mg.Masters.Items[mIndex2]
		return
	}

	// Swap master with slave
	if mIndex1 >= 0 && sIndex2 >= 0 {
		mg.Slaves.Items[sIndex2], mg.Masters.Items[mIndex1] = mg.Masters.Items[mIndex1], mg.Slaves.Items[sIndex2]
		return
	}

	// Swap slave with master
	if sIndex1 >= 0 && mIndex2 >= 0 {
		mg.Masters.Items[mIndex2], mg.Slaves.Items[sIndex1] = mg.Slaves.Items[sIndex1], mg.Masters.Items[mIndex2]
		return
	}

	// Swap slave with slave
	if sIndex1 >= 0 && sIndex2 >= 0 {
		mg.Slaves.Items[sIndex2], mg.Slaves.Items[sIndex1] = mg.Slaves.Items[sIndex1], mg.Slaves.Items[sIndex2]
		return
	}
}

func (mg *Manager) NextClient() *Client {
	clients := mg.Clients(Stacked)
	last := len(clients) - 1

	// Get next window
	next := -1
	for i, c := range clients {
		if c.Win.Id == ActiveWindow {
			next = i + 1
			if next > last {
				next = 0
			}
			break
		}
	}

	// Invalid active window
	if next == -1 {
		return nil
	}

	return clients[next]
}

func (mg *Manager) PreviousClient() *Client {
	clients := mg.Clients(Stacked)
	last := len(clients) - 1

	// Get previous window
	prev := -1
	for i, c := range clients {
		if c.Win.Id == ActiveWindow {
			prev = i - 1
			if prev < 0 {
				prev = last
			}
			break
		}
	}

	// Invalid active window
	if prev == -1 {
		return nil
	}

	return clients[prev]
}

func (mg *Manager) IncreaseMaster() {

	// Increase master area
	if len(mg.Slaves.Items) > 1 && mg.Masters.MaxAllowed < common.Config.WindowMastersMax {
		mg.Masters.MaxAllowed += 1
		mg.updateMasters(append(mg.Masters.Items, mg.Slaves.Items[0]))
		mg.updateSlaves(mg.Slaves.Items[1:])
	}

	log.Info("Increase masters to ", mg.Masters.MaxAllowed)
}

func (mg *Manager) DecreaseMaster() {

	// Decrease master area
	if len(mg.Masters.Items) > 0 {
		mg.Masters.MaxAllowed -= 1
		mg.updateSlaves(append([]*Client{mg.Masters.Items[len(mg.Masters.Items)-1]}, mg.Slaves.Items...))
		mg.updateMasters(mg.Masters.Items[:len(mg.Masters.Items)-1])
	}

	log.Info("Decrease masters to ", mg.Masters.MaxAllowed)
}

func (mg *Manager) IncreaseSlave() {

	// Increase slave area
	if mg.Slaves.MaxAllowed < common.Config.WindowSlavesMax {
		mg.Slaves.MaxAllowed += 1
		mg.updateSlaves(mg.Slaves.Items)
	}

	log.Info("Increase slaves to ", mg.Slaves.MaxAllowed)
}

func (mg *Manager) DecreaseSlave() {

	// Decrease slave area
	if mg.Slaves.MaxAllowed > 1 {
		mg.Slaves.MaxAllowed -= 1
		mg.updateSlaves(mg.Slaves.Items)
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
	if i == j || i < 0 || i >= len(ps) || j < 0 || j >= len(ps) {
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

func (mg *Manager) Index(windows *Clients, c *Client) int {

	// Traverse client list
	for i, m := range windows.Items {
		if m.Win.Id == c.Win.Id {
			return i
		}
	}

	return -1
}

func (mg *Manager) Ordered(windows *Clients) []*Client {
	ordered := []*Client{}

	// Create ordered client list
	for _, w := range Windows {
		for _, c := range windows.Items {
			if w == c.Win.Id {
				ordered = append(ordered, c)
				break
			}
		}
	}

	return ordered
}

func (mg *Manager) Visible(windows *Clients) []*Client {
	visible := make([]*Client, int(math.Min(float64(len(windows.Items)), float64(windows.MaxAllowed))))

	// Create visible client list
	for _, c := range mg.Ordered(windows) {
		visible[mg.Index(windows, c)%windows.MaxAllowed] = c
	}

	return visible
}

func (mg *Manager) Clients(flag uint8) []*Client {
	switch flag {
	case Stacked:
		return append(mg.Masters.Items, mg.Slaves.Items...)
	case Visible:
		return append(mg.Visible(mg.Masters), mg.Visible(mg.Slaves)...)
	default:
		return make([]*Client, 0)
	}
}

func (mg *Manager) updateMasters(cs []*Client) {
	mg.Masters.Items = mg.Ordered(&Clients{Items: cs})
	mg.Proportions.MasterMaster = calcProportions(int(math.Min(float64(len(mg.Masters.Items)), float64(mg.Masters.MaxAllowed))))
}

func (mg *Manager) updateSlaves(cs []*Client) {
	mg.Slaves.Items = mg.Ordered(&Clients{Items: cs})
	mg.Proportions.SlaveSlave = calcProportions(int(math.Min(float64(len(mg.Slaves.Items)), float64(mg.Slaves.MaxAllowed))))
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

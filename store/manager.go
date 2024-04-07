package store

import (
	"fmt"
	"math"

	"github.com/leukipp/cortile/v2/common"

	log "github.com/sirupsen/logrus"
)

type Manager struct {
	Name        string       // Manager name with window clients
	Location    *Location    // Manager workspace and screen location
	Proportions *Proportions // Manager proportions of window clients
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
	MasterSlave  map[int][]float64 // Master-slave proportions
	MasterMaster map[int][]float64 // Master-master proportions
	SlaveSlave   map[int][]float64 // Slave-slave proportions
}

type Clients struct {
	Stacked    []*Client `json:"-"` // List of stored window clients
	MaxAllowed int       // Currently maximum allowed clients
}

const (
	Stacked uint8 = 1 // Flag for stacked (all) clients
	Visible uint8 = 2 // Flag for visible (top) clients
)

func CreateManager(loc Location) *Manager {
	return &Manager{
		Name:     fmt.Sprintf("manager-%d-%d", loc.DeskNum, loc.ScreenNum),
		Location: &loc,
		Proportions: &Proportions{
			MasterSlave:  calcProportions(2),
			MasterMaster: calcProportions(common.Config.WindowMastersMax),
			SlaveSlave:   calcProportions(common.Config.WindowSlavesMax),
		},
		Masters: &Clients{
			Stacked:    make([]*Client, 0),
			MaxAllowed: 1,
		},
		Slaves: &Clients{
			Stacked:    make([]*Client, 0),
			MaxAllowed: common.Config.WindowSlavesMax,
		},
	}
}

func (mg *Manager) AddClient(c *Client) {
	if mg.IsMaster(c) || mg.IsSlave(c) {
		return
	}

	log.Debug("Add client for manager [", c.Latest.Class, ", ", mg.Name, "]")

	// Fill up master area then slave area
	if len(mg.Masters.Stacked) < mg.Masters.MaxAllowed {
		mg.Masters.Stacked = addClient(mg.Masters.Stacked, c)
	} else {
		mg.Slaves.Stacked = addClient(mg.Slaves.Stacked, c)
	}
}

func (mg *Manager) RemoveClient(c *Client) {
	log.Debug("Remove client from manager [", c.Latest.Class, ", ", mg.Name, "]")

	// Remove master window
	mi := mg.Index(mg.Masters, c)
	if mi >= 0 {
		if len(mg.Slaves.Stacked) > 0 {
			mg.SwapClient(mg.Masters.Stacked[mi], mg.Slaves.Stacked[0])
			mg.Slaves.Stacked = mg.Slaves.Stacked[1:]
		} else {
			mg.Masters.Stacked = removeClient(mg.Masters.Stacked, mi)
		}
	}

	// Remove slave window
	si := mg.Index(mg.Slaves, c)
	if si >= 0 {
		mg.Slaves.Stacked = removeClient(mg.Slaves.Stacked, si)
	}
}

func (mg *Manager) MakeMaster(c *Client) {
	log.Info("Make window master [", c.Latest.Class, ", ", mg.Name, "]")

	// Swap window with first master
	if len(mg.Masters.Stacked) > 0 {
		mg.SwapClient(c, mg.Masters.Stacked[0])
	}
}

func (mg *Manager) SwapClient(c1 *Client, c2 *Client) {
	log.Info("Swap clients [", c1.Latest.Class, "-", c2.Latest.Class, ", ", mg.Name, "]")

	mIndex1 := mg.Index(mg.Masters, c1)
	sIndex1 := mg.Index(mg.Slaves, c1)

	mIndex2 := mg.Index(mg.Masters, c2)
	sIndex2 := mg.Index(mg.Slaves, c2)

	// Swap master with master
	if mIndex1 >= 0 && mIndex2 >= 0 {
		mg.Masters.Stacked[mIndex2], mg.Masters.Stacked[mIndex1] = mg.Masters.Stacked[mIndex1], mg.Masters.Stacked[mIndex2]
		return
	}

	// Swap master with slave
	if mIndex1 >= 0 && sIndex2 >= 0 {
		mg.Slaves.Stacked[sIndex2], mg.Masters.Stacked[mIndex1] = mg.Masters.Stacked[mIndex1], mg.Slaves.Stacked[sIndex2]
		return
	}

	// Swap slave with master
	if sIndex1 >= 0 && mIndex2 >= 0 {
		mg.Masters.Stacked[mIndex2], mg.Slaves.Stacked[sIndex1] = mg.Slaves.Stacked[sIndex1], mg.Masters.Stacked[mIndex2]
		return
	}

	// Swap slave with slave
	if sIndex1 >= 0 && sIndex2 >= 0 {
		mg.Slaves.Stacked[sIndex2], mg.Slaves.Stacked[sIndex1] = mg.Slaves.Stacked[sIndex1], mg.Slaves.Stacked[sIndex2]
		return
	}
}

func (mg *Manager) NextClient() *Client {
	clients := mg.Clients(Stacked)
	last := len(clients) - 1

	// Get next window
	next := -1
	for i, c := range clients {
		if c.Win.Id == Windows.Active {
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
		if c.Win.Id == Windows.Active {
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
	if len(mg.Slaves.Stacked) > 1 && mg.Masters.MaxAllowed < common.Config.WindowMastersMax {
		mg.Masters.MaxAllowed += 1
		mg.Masters.Stacked = append(mg.Masters.Stacked, mg.Slaves.Stacked[0])
		mg.Slaves.Stacked = mg.Slaves.Stacked[1:]
	}

	log.Info("Increase masters to ", mg.Masters.MaxAllowed)
}

func (mg *Manager) DecreaseMaster() {

	// Decrease master area
	if len(mg.Masters.Stacked) > 0 {
		mg.Masters.MaxAllowed -= 1
		mg.Slaves.Stacked = append([]*Client{mg.Masters.Stacked[len(mg.Masters.Stacked)-1]}, mg.Slaves.Stacked...)
		mg.Masters.Stacked = mg.Masters.Stacked[:len(mg.Masters.Stacked)-1]
	}

	log.Info("Decrease masters to ", mg.Masters.MaxAllowed)
}

func (mg *Manager) IncreaseSlave() {

	// Increase slave area
	if mg.Slaves.MaxAllowed < common.Config.WindowSlavesMax {
		mg.Slaves.MaxAllowed += 1
	}

	log.Info("Increase slaves to ", mg.Slaves.MaxAllowed)
}

func (mg *Manager) DecreaseSlave() {

	// Decrease slave area
	if mg.Slaves.MaxAllowed > 1 {
		mg.Slaves.MaxAllowed -= 1
	}

	log.Info("Decrease slaves to ", mg.Slaves.MaxAllowed)
}

func (mg *Manager) IncreaseProportion() {
	precision := 1.0 / common.Config.ProportionStep
	proportion := math.Round(mg.Proportions.MasterSlave[2][0]*precision)/precision + common.Config.ProportionStep

	// Increase root proportion
	mg.SetProportions(mg.Proportions.MasterSlave[2], proportion, 0, 1)
}

func (mg *Manager) DecreaseProportion() {
	precision := 1.0 / common.Config.ProportionStep
	proportion := math.Round(mg.Proportions.MasterSlave[2][0]*precision)/precision - common.Config.ProportionStep

	// Decrease root proportion
	mg.SetProportions(mg.Proportions.MasterSlave[2], proportion, 0, 1)
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
	for i, m := range windows.Stacked {
		if m.Win.Id == c.Win.Id {
			return i
		}
	}

	return -1
}

func (mg *Manager) Ordered(windows *Clients) []*Client {
	ordered := []*Client{}

	// Create ordered client list
	for _, w := range Windows.Stacked {
		for _, c := range windows.Stacked {
			if w == c.Win.Id {
				ordered = append(ordered, c)
				break
			}
		}
	}

	return ordered
}

func (mg *Manager) Visible(windows *Clients) []*Client {
	visible := make([]*Client, int(math.Min(float64(len(windows.Stacked)), float64(windows.MaxAllowed))))

	// Create visible client list
	for _, c := range mg.Ordered(windows) {
		visible[mg.Index(windows, c)%windows.MaxAllowed] = c
	}

	return visible
}

func (mg *Manager) Clients(flag uint8) []*Client {
	switch flag {
	case Stacked:
		return append(mg.Masters.Stacked, mg.Slaves.Stacked...)
	case Visible:
		return append(mg.Visible(mg.Masters), mg.Visible(mg.Slaves)...)
	}
	return make([]*Client, 0)
}

func addClient(cs []*Client, c *Client) []*Client {
	return append([]*Client{c}, cs...)
}

func removeClient(cs []*Client, i int) []*Client {
	return append(cs[:i], cs[i+1:]...)
}

func calcProportions(n int) map[int][]float64 {
	p := map[int][]float64{}
	for i := 1; i <= n; i++ {
		for j := 1; j <= i; j++ {
			p[i] = append(p[i], 1.0/float64(i))
		}
	}
	return p
}

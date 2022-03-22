package store

import (
	"github.com/leukipp/cortile/common"

	log "github.com/sirupsen/logrus"
)

type Manager struct {
	Masters        []Client // List of master window clients
	Slaves         []Client // List of slave window clients
	AllowedMasters int      // Number of maximal allowed masters
}

func CreateManager() *Manager {
	return &Manager{
		Masters:        make([]Client, 0),
		Slaves:         make([]Client, 0),
		AllowedMasters: 1,
	}
}

func (mg *Manager) Add(c *Client) {
	log.Info("Add client [", c.Class, "]")

	// Fill up master area then slave area
	if len(mg.Masters) < mg.AllowedMasters {
		mg.Masters = addClient(mg.Masters, c)
	} else {
		mg.Slaves = addClient(mg.Slaves, c)
	}
}

func (mg *Manager) Remove(c *Client) {
	if c.Win == nil {
		return
	}

	log.Info("Remove client [", c.Class, "]")

	// Remove master window
	mi := getIndex(mg.Masters, c)
	if mi >= 0 {
		if len(mg.Slaves) > 0 {
			mg.Masters[mi] = mg.Slaves[0]
			mg.Slaves = mg.Slaves[1:]
		} else {
			mg.Masters = removeClient(mg.Masters, mi)
		}
		return
	}

	// Remove slave window
	si := getIndex(mg.Slaves, c)
	if si >= 0 {
		mg.Slaves = removeClient(mg.Slaves, si)
		return
	}
}

func (mg *Manager) IncreaseMaster() {

	// Increase master area
	if len(mg.Slaves) > 1 {
		mg.AllowedMasters = mg.AllowedMasters + 1
		mg.Masters = append(mg.Masters, mg.Slaves[0])
		mg.Slaves = mg.Slaves[1:]
	}

	log.Info("Increase masters to ", mg.AllowedMasters)
}

func (mg *Manager) DecreaseMaster() {

	// Decrease master area
	if len(mg.Masters) > 1 {
		mg.AllowedMasters = mg.AllowedMasters - 1
		mg.Slaves = append([]Client{mg.Masters[len(mg.Masters)-1]}, mg.Slaves...)
		mg.Masters = mg.Masters[:len(mg.Masters)-1]
	}

	log.Info("Decrease masters to ", mg.AllowedMasters)
}

func (mg *Manager) MakeMaster(c *Client) {
	if c.Win == nil {
		return
	}

	log.Info("Make window master [", c.Class, "]")

	// Swap master with master
	mi := getIndex(mg.Masters, c)
	if mi >= 0 {
		mg.Masters[0], mg.Masters[mi] = mg.Masters[mi], mg.Masters[0]
		return
	}

	// Swap slave with master
	si := getIndex(mg.Slaves, c)
	if si >= 0 {
		mg.Masters[0], mg.Slaves[si] = mg.Slaves[si], mg.Masters[0]
		return
	}
}

func (mg *Manager) IsMaster(c *Client) bool {

	// Check if window is master
	return getIndex(mg.Masters, c) >= 0
}

func (mg *Manager) Next() Client {
	clients := mg.Clients()
	lastIndex := len(clients) - 1

	// Get next window
	for i, c := range clients {
		if c.Win.Id == common.ActiveWin {
			next := i + 1
			if next > lastIndex {
				next = 0
			}
			return clients[next]
		}
	}

	return Client{}
}

func (mg *Manager) Previous() Client {
	clients := mg.Clients()
	lastIndex := len(clients) - 1

	// Get previous window
	for i, c := range clients {
		if c.Win.Id == common.ActiveWin {
			prev := i - 1
			if prev < 0 {
				prev = lastIndex
			}
			return clients[prev]
		}
	}

	return Client{}
}

func (mg *Manager) Clients() []Client {
	return append(mg.Masters, mg.Slaves...)
}

func addClient(cs []Client, c *Client) []Client {
	return append([]Client{*c}, cs...)
}

func removeClient(cs []Client, i int) []Client {
	return append(cs[:i], cs[i+1:]...)
}

func getIndex(cs []Client, c *Client) int {
	for i, m := range cs {
		if m.Win.Id == c.Win.Id {
			return i
		}
	}
	return -1
}

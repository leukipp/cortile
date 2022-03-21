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

func (st *Manager) Add(c *Client) {
	log.Info("Add client [", c.Class, "]")

	// Fill up master area then slave area
	if len(st.Masters) < st.AllowedMasters {
		st.Masters = addClient(st.Masters, c)
	} else {
		st.Slaves = addClient(st.Slaves, c)
	}
}

func (st *Manager) Remove(c *Client) {
	if c.Win == nil {
		return
	}

	log.Info("Remove client [", c.Class, "]")

	// Remove master window
	mi := getIndex(st.Masters, c)
	if mi >= 0 {
		if len(st.Slaves) > 0 {
			st.Masters[mi] = st.Slaves[0]
			st.Slaves = st.Slaves[1:]
		} else {
			st.Masters = removeClient(st.Masters, mi)
		}
		return
	}

	// Remove slave window
	si := getIndex(st.Slaves, c)
	if si >= 0 {
		st.Slaves = removeClient(st.Slaves, si)
		return
	}
}

func (st *Manager) IncreaseMaster() {

	// Increase master area
	if len(st.Slaves) > 1 {
		st.AllowedMasters = st.AllowedMasters + 1
		st.Masters = append(st.Masters, st.Slaves[0])
		st.Slaves = st.Slaves[1:]
	}

	log.Info("Increase masters to ", st.AllowedMasters)
}

func (st *Manager) DecreaseMaster() {

	// Decrease master area
	if len(st.Masters) > 1 {
		st.AllowedMasters = st.AllowedMasters - 1
		st.Slaves = append([]Client{st.Masters[len(st.Masters)-1]}, st.Slaves...)
		st.Masters = st.Masters[:len(st.Masters)-1]
	}

	log.Info("Decrease masters to ", st.AllowedMasters)
}

func (st *Manager) MakeMaster(c *Client) {
	if c.Win == nil {
		return
	}

	log.Info("Make window master [", c.Class, "]")

	// Swap master with master
	mi := getIndex(st.Masters, c)
	if mi >= 0 {
		st.Masters[0], st.Masters[mi] = st.Masters[mi], st.Masters[0]
		return
	}

	// Swap slave with master
	si := getIndex(st.Slaves, c)
	if si >= 0 {
		st.Masters[0], st.Slaves[si] = st.Slaves[si], st.Masters[0]
		return
	}
}

func (st *Manager) Next() Client {
	clients := st.Clients()
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

func (st *Manager) Previous() Client {
	clients := st.Clients()
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

func (st *Manager) Clients() []Client {
	return append(st.Masters, st.Slaves...)
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

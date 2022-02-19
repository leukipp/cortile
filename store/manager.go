package store

import (
	"github.com/leukipp/Cortile/common"

	log "github.com/sirupsen/logrus"
)

type Manager struct {
	Masters        []Client
	Slaves         []Client
	AllowedMasters int
}

func CreateManager() *Manager {
	return &Manager{
		Masters:        make([]Client, 0),
		Slaves:         make([]Client, 0),
		AllowedMasters: 1,
	}
}

func (st *Manager) Add(c Client) {
	if len(st.Masters) < st.AllowedMasters {
		st.Masters = addItem(st.Masters, c)
	} else {
		st.Slaves = addItem(st.Slaves, c)
	}
}

func addItem(cs []Client, c Client) []Client {
	return append([]Client{c}, cs...)
}

func (st *Manager) Remove(c Client) {

	log.Error("--- len masters 1 --- : ", len(st.Masters), ", id=", c.Win.Id)
	log.Error("--- len slaves 1 --- : ", len(st.Slaves), ", id=", c.Win.Id)

	for i, m := range st.Masters {
		if m.Win.Id == c.Win.Id {
			if len(st.Slaves) > 0 {
				log.Error("--- remove 1 --- : ", c)
				st.Masters[i] = st.Slaves[0]
				st.Slaves = st.Slaves[1:]
			} else {
				log.Error("--- remove 2 --- : ", c)
				st.Masters = removeItem(st.Masters, i)
			}

			log.Error("--- len masters 2 --- : ", len(st.Masters))
			log.Error("--- len slaves 2 --- : ", len(st.Slaves))

			return
		}
	}

	for i, s := range st.Slaves {
		if s.Win.Id == c.Win.Id {
			log.Error("--- remove 3 --- : ", c)
			st.Slaves = removeItem(st.Slaves, i)
			log.Error("--- len masters 3 --- : ", len(st.Masters))
			log.Error("--- len slaves 3 --- : ", len(st.Slaves))
			return
		}
	}
}

func removeItem(cs []Client, i int) []Client {
	return append(cs[:i], cs[i+1:]...)
}

func (st *Manager) IncreaseMaster() {
	if len(st.Slaves) > 1 {
		st.AllowedMasters = st.AllowedMasters + 1
		st.Masters = append(st.Masters, st.Slaves[0])
		st.Slaves = st.Slaves[1:]
	}

	log.Info("Increase masters to ", st.AllowedMasters)
}

func (st *Manager) DecreaseMaster() {
	if len(st.Masters) > 1 {
		st.AllowedMasters = st.AllowedMasters - 1
		mlen := len(st.Masters)
		st.Slaves = append([]Client{st.Masters[mlen-1]}, st.Slaves...)
		st.Masters = st.Masters[:mlen-1]
	}

	log.Info("Decrease masters to ", st.AllowedMasters)
}

func (st *Manager) MakeMaster(c Client) {
	if c.Win == nil {
		return
	}

	log.Info("Make window master [", c.Class, "]")

	for i, master := range st.Masters {
		if master.Win.Id == c.Win.Id {
			st.Masters[0], st.Masters[i] = st.Masters[i], st.Masters[0]
		}
	}

	for i, slave := range st.Slaves {
		if slave.Win.Id == c.Win.Id {
			st.Masters[0], st.Slaves[i] = st.Slaves[i], st.Masters[0]
		}
	}
}

func (st *Manager) All() []Client {
	return append(st.Masters, st.Slaves...)
}

func (st *Manager) Next() Client {
	clients := st.All()
	lastIndex := len(clients) - 1

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
	clients := st.All()
	lastIndex := len(clients) - 1

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

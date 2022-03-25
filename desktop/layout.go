package desktop

import "github.com/leukipp/cortile/store"

type Layout interface {
	Do()
	Undo()
	AddClient(c *store.Client)
	RemoveClient(c *store.Client)
	MakeMaster(c *store.Client)
	SwapClient(c1 *store.Client, c2 *store.Client)
	NextClient()
	PreviousClient()
	IncreaseMaster()
	DecreaseMaster()
	IncrementProportion()
	DecrementProportion()
	SetProportion(p float64)
	GetType() string
	GetManager() *store.Manager
}

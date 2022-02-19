package desktop

import "github.com/leukipp/Cortile/store"

type Layout interface {
	Do()
	Undo()
	Add(c store.Client)
	Remove(c store.Client)
	MakeMaster(c store.Client)
	IncreaseMaster()
	DecreaseMaster()
	NextClient()
	PreviousClient()
	IncrementProportion()
	DecrementProportion()
	SetProportion(p float64)
	GetType() string
	GetManager() *store.Manager
}

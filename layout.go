package main

const (
	MASTER_MAX_PROPORTION = 0.9
	MASTER_MIN_PROPORTION = 0.1
)

type Layout interface {
	Do()
	Undo()
	Add(c Client)
	Remove(c Client)
	MakeMaster(c Client)
	IncMaster()
	DecreaseMaster()
	NextClient()
	PreviousClient()
	IncrementMaster()
	DecrementMaster()
	sto() *Store
}

type VertHorz struct {
	*Store
	Proportion   float64
	WorkspaceNum uint
}

func (l *VertHorz) Undo() {
	for _, c := range append(l.masters, l.slaves...) {
		c.Restore()
	}
}

func (l *VertHorz) NextClient() {
	c := l.Next()
	c.Activate()
}

func (l *VertHorz) PreviousClient() {
	c := l.Previous()
	c.Activate()
}

func (l *VertHorz) IncrementMaster() {
	value := l.Proportion + Config.Proportion
	if value >= MASTER_MAX_PROPORTION {
		return
	}
	l.Proportion = value
}

func (l *VertHorz) DecrementMaster() {
	value := l.Proportion - Config.Proportion
	if value <= MASTER_MIN_PROPORTION {
		return
	}
	l.Proportion = value
}

func (l *VertHorz) sto() *Store {
	return l.Store
}

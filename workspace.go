package main

import (
	"fmt"

	"github.com/blrsn/zentile/state"
)

type Workspace struct {
	IsTiling        bool
	activeLayoutNum uint
	layouts         []Layout
}

func CreateWorkspaces() map[uint]*Workspace {
	workspaces := make(map[uint]*Workspace)
	for i := uint(0); i < state.DeskCount; i++ {
		ws := Workspace{
			IsTiling: false,
			layouts:  createLayouts(i),
		}

		workspaces[i] = &ws
	}

	return workspaces
}

func createLayouts(workspaceNum uint) []Layout {
	return []Layout{
		&VerticalLayout{&VertHorz{
			Store:        buildStore(),
			Proportion:   0.5,
			WorkspaceNum: workspaceNum,
		}},
		&HorizontalLayout{&VertHorz{
			Store:        buildStore(),
			Proportion:   0.5,
			WorkspaceNum: workspaceNum,
		}},
		&FullScreen{
			Store:        buildStore(),
			WorkspaceNum: workspaceNum,
		},
	}
}

func (ws *Workspace) ActiveLayout() Layout {
	return ws.layouts[ws.activeLayoutNum]
}

// Cycle through the available layouts
func (ws *Workspace) SwitchLayout() {
	ws.activeLayoutNum = (ws.activeLayoutNum + 1) % uint(len(ws.layouts))
	ws.ActiveLayout().Do()
}

// Adds client to all the layouts in a workspace
func (ws *Workspace) AddClient(c Client) {
	for _, l := range ws.layouts {
		l.Add(c)
	}
}

// Removes client from all the layouts in a workspace
func (ws *Workspace) RemoveClient(c Client) {
	for _, l := range ws.layouts {
		l.Remove(c)
	}
}

// Tiles the active layout in a workspace
func (ws *Workspace) Tile() {
	if ws.IsTiling {
		ws.ActiveLayout().Do()
	}
}

// Untiles the active layout in a workspace.
func (ws *Workspace) Untile() {
	ws.IsTiling = false
	ws.ActiveLayout().Undo()
}

func (ws *Workspace) printStore() {
	l := ws.ActiveLayout()
	st := l.sto()
	fmt.Println("Number of masters is ", len(st.masters))
	fmt.Println("Number of slaves is", len(st.slaves))

	for i, c := range st.masters {
		fmt.Println("master ", " ", i, " - ", c.name())
	}

	for i, c := range st.slaves {
		fmt.Println("slave ", " ", i, " - ", c.name())
	}
}

package main

import (
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/blrsn/zentile/state"
)

type tracker struct {
	clients    map[xproto.Window]Client // List of clients that are being tracked.
	workspaces map[uint]*Workspace
}

func initTracker(ws map[uint]*Workspace) *tracker {
	t := tracker{
		clients:    make(map[xproto.Window]Client),
		workspaces: ws,
	}

	xevent.PropertyNotifyFun(t.handleClientUpdates).Connect(state.X, state.X.RootWin())
	t.populateClients()
	return &t
}

// UpdateClients updates the list of tracked clients with the most up to date list of clients.
func (tr *tracker) populateClients() {
	clientList, _ := ewmh.ClientListStackingGet(state.X)
	for _, w := range clientList {
		if isHidden(w) || shouldIgnore(w) {
			continue
		}
		tr.trackWindow(w)
	}

	// If window is tracked, but not in client list
	for wid := range tr.clients {
		found := false
		for _, w := range clientList {
			if w == wid {
				found = true
				break
			}
		}

		if !found {
			tr.unTrack(wid)
		}
	}
}

func (tr *tracker) IsTracked(w xproto.Window) bool {
	_, ok := tr.clients[w]
	return ok
}

// Adds window to Tracked Clients and layouts.
func (tr *tracker) trackWindow(w xproto.Window) {
	if tr.IsTracked(w) {
		return
	}

	c := newClient(w)
	if c.Desk > state.DeskCount {
		return
	}
	tr.attachHandlers(&c)

	tr.clients[c.window.Id] = c
	ws := tr.workspaces[c.Desk]
	ws.AddClient(c)

}

func (tr *tracker) unTrack(w xproto.Window) {
	c, ok := tr.clients[w]
	if ok {
		ws := tr.workspaces[c.Desk]
		ws.RemoveClient(c)
		xevent.Detach(state.X, w)
		delete(tr.clients, w)
	}
}

func (tr *tracker) handleClientUpdates(X *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
	if aname, _ := xprop.AtomName(state.X, ev.Atom); aname != "_NET_CLIENT_LIST_STACKING" {
		return
	}

	tr.populateClients()
	tr.workspaces[state.CurrentDesk].Tile()
}

func (tr *tracker) handleMinimizedClient(c *Client) {
	states, _ := ewmh.WmStateGet(state.X, c.window.Id)
	for _, state := range states {
		if state == "_NET_WM_STATE_HIDDEN" {
			tr.workspaces[c.Desk].RemoveClient(*c)
			tr.unTrack(c.window.Id)
			tr.workspaces[c.Desk].Tile()
		}
	}
}

func (tr *tracker) handleDesktopChange(c *Client) {
	newDesk, _ := ewmh.WmDesktopGet(state.X, c.window.Id)
	oldDesk := c.Desk

	tr.workspaces[oldDesk].RemoveClient(*c)
	tr.workspaces[newDesk].AddClient(*c)

	c.Desk = newDesk
	if tr.workspaces[oldDesk].IsTiling {
		tr.workspaces[oldDesk].Tile()
	}

	if tr.workspaces[newDesk].IsTiling {
		tr.workspaces[newDesk].Tile()
	} else {
		c.Restore()
	}
}

func (tr *tracker) attachHandlers(c *Client) {
	c.window.Listen(xproto.EventMaskPropertyChange)

	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		if aname, _ := xprop.AtomName(state.X, ev.Atom); aname == "_NET_WM_STATE" {
			tr.handleMinimizedClient(c)
		}
	}).Connect(state.X, c.window.Id)

	xevent.PropertyNotifyFun(func(x *xgbutil.XUtil, ev xevent.PropertyNotifyEvent) {
		if aname, _ := xprop.AtomName(state.X, ev.Atom); aname == "_NET_WM_DESKTOP" {
			tr.handleDesktopChange(c)
		}
	}).Connect(state.X, c.window.Id)
}

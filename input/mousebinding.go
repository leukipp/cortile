package input

import (
	"time"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/desktop"
	"github.com/leukipp/cortile/v2/store"
	"github.com/leukipp/cortile/v2/ui"

	log "github.com/sirupsen/logrus"
)

var (
	workspace *desktop.Workspace // Stores previous workspace (for comparison only)
	pointer   *store.XPointer    // Stores previous pointer (for comparison only)
	hover     *time.Timer        // Timer to delay hover events
)

func BindMouse(tr *desktop.Tracker) {
	poll(100, func() {
		store.PointerUpdate(store.X)

		// Reset tracker handler
		resetTracker(tr)

		// Evaluate workspace state
		updateWorkspace(tr)

		// Evaluate corner state
		updateCorner(tr)

		// Evaluate focus state
		updateFocus(tr)

		// Store last pointer
		pointer = store.Pointer
	})
}

func resetTracker(tr *desktop.Tracker) {
	if pointer == nil || pointer.Position != store.Pointer.Position {
		return
	}

	// Reset tracker handler
	if !tr.Handlers.MoveClient.Active() {
		tr.Handlers.Reset()
	}
}

func updateWorkspace(tr *desktop.Tracker) {
	ws := tr.ActiveWorkspace()
	if ws == nil || ws == workspace {
		return
	}
	log.Info("Active workspace updated [", ws.Name, "]")

	// Communicate workplace change
	tr.Channels.Event <- "workplace_change"

	// Update systray icon
	ui.UpdateIcon(ws)

	// Store last workspace
	workspace = ws
}

func updateCorner(tr *desktop.Tracker) {
	hc := store.HotCorner()
	if hc == nil {
		return
	}

	// Communicate corner change
	tr.Channels.Event <- "corner_change"

	// Execute action
	ExecuteAction(common.Config.Corners[hc.Name], tr, tr.ActiveWorkspace())
}

func updateFocus(tr *desktop.Tracker) {
	ws := tr.ActiveWorkspace()
	if ws == nil || pointer == nil || hover != nil {
		return
	}

	// Ignore stationary pointer position
	if pointer.Position == store.Pointer.Position {
		return
	}

	// Ignore untracked clients
	active := tr.ActiveClient()
	hovered := tr.ClientAt(ws, store.Pointer.Position)
	if active == nil || hovered == nil {
		return
	}
	log.Info("Hovered window updated [", hovered.Latest.Class, "]")

	// Delay hover event by given duration
	if common.Config.WindowFocusDelay == 0 {
		return
	}
	hover = time.AfterFunc(time.Duration(common.Config.WindowFocusDelay)*time.Millisecond, func() {
		hover = nil

		// Hovered client window has changed in the meantime
		if hovered != tr.ClientAt(ws, store.Pointer.Position) {
			return
		}

		// Focus hovered client window
		if hovered != active && ws.TilingEnabled() && !tr.Handlers.Active() {
			store.ActiveWindowSet(store.X, hovered.Window)
		}
	})
}

func poll(t time.Duration, fun func()) {
	go func() {
		for range time.Tick(t * time.Millisecond) {
			fun()
		}
	}()
}

package input

import (
	"os"
	"strings"

	"os/exec"

	"github.com/jezek/xgbutil/xevent"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/desktop"
	"github.com/leukipp/cortile/v2/store"
	"github.com/leukipp/cortile/v2/ui"

	log "github.com/sirupsen/logrus"
)

var (
	executeCallbacksFun []func(string, uint, uint) // Execute events callback functions
)

func ExecuteAction(action string, tr *desktop.Tracker, ws *desktop.Workspace) bool {
	success := false
	if tr == nil || ws == nil {
		return false
	}

	log.Info("Execute action ", action, " [", ws.Name, "]")

	// Choose action command
	switch action {
	case "":
		success = false
	case "enable":
		success = EnableTiling(tr, ws)
	case "disable":
		success = DisableTiling(tr, ws)
	case "toggle":
		success = ToggleTiling(tr, ws)
	case "decoration":
		success = ToggleDecoration(tr, ws)
	case "restore":
		success = Restore(tr, ws)
	case "cycle_next":
		success = CycleNext(tr, ws)
	case "cycle_previous":
		success = CyclePrevious(tr, ws)
	case "layout_vertical_left":
		success = VerticalLeftLayout(tr, ws)
	case "layout_vertical_right":
		success = VerticalRightLayout(tr, ws)
	case "layout_horizontal_top":
		success = HorizontalTopLayout(tr, ws)
	case "layout_horizontal_bottom":
		success = HorizontalBottomLayout(tr, ws)
	case "layout_maximized":
		success = MaximizedLayout(tr, ws)
	case "layout_fullscreen":
		success = FullscreenLayout(tr, ws)
	case "master_make":
		success = MakeMaster(tr, ws)
	case "master_make_next":
		success = MakeMasterNext(tr, ws)
	case "master_make_previous":
		success = MakeMasterPrevious(tr, ws)
	case "master_increase":
		success = IncreaseMaster(tr, ws)
	case "master_decrease":
		success = DecreaseMaster(tr, ws)
	case "slave_increase":
		success = IncreaseSlave(tr, ws)
	case "slave_decrease":
		success = DecreaseSlave(tr, ws)
	case "proportion_increase":
		success = IncreaseProportion(tr, ws)
	case "proportion_decrease":
		success = DecreaseProportion(tr, ws)
	case "window_next":
		success = NextWindow(tr, ws)
	case "window_previous":
		success = PreviousWindow(tr, ws)
	case "reset":
		success = Reset(tr, ws)
	case "exit":
		success = Exit(tr)
	default:
		success = External(action)
	}

	// Check success
	if !success {
		return false
	}

	// Execute callbacks
	executeCallbacks(action, ws.Location.DeskNum, ws.Location.ScreenNum)

	return true
}

func ExecuteActions(action string, tr *desktop.Tracker, mod string) bool {
	results := []bool{}

	active := tr.ActiveWorkspace()
	for _, ws := range tr.Workspaces {

		// Execute only on active screen
		if mod == "current" && ws.Location != active.Location {
			continue
		}

		// Execute only on active workspace
		if mod == "screens" && (ws.Location.DeskNum != active.Location.DeskNum) {
			continue
		}

		// Execute action and store results
		success := ExecuteAction(action, tr, ws)
		results = append(results, success)
	}

	return common.AllTrue(results)
}

func EnableTiling(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	ws.EnableTiling()
	tr.Update()
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func DisableTiling(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	ws.DisableTiling()
	tr.Restore(ws, store.Latest)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func ToggleTiling(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return EnableTiling(tr, ws)
	}
	return DisableTiling(tr, ws)
}

func EnableDecoration(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	mg := ws.ActiveLayout().GetManager()

	mg.EnableDecoration()
	tr.Update()
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func DisableDecoration(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	mg := ws.ActiveLayout().GetManager()

	mg.DisableDecoration()
	tr.Update()
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func ToggleDecoration(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	mg := ws.ActiveLayout().GetManager()
	if mg.DecorationDisabled() {
		return EnableDecoration(tr, ws)
	}
	return DisableDecoration(tr, ws)
}

func Restore(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	ws.DisableTiling()
	tr.Restore(ws, store.Original)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func CycleNext(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	if int(ws.ActiveLayoutNum) == len(ws.Layouts)-2 {
		ws.CycleLayout(2)
	} else {
		ws.CycleLayout(1)
	}
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func CyclePrevious(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	if int(ws.ActiveLayoutNum) == 0 {
		ws.CycleLayout(-2)
	} else {
		ws.CycleLayout(-1)
	}
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func VerticalLeftLayout(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "vertical-left" {
			ws.SetLayout(uint(i))
		}
	}
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func VerticalRightLayout(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "vertical-right" {
			ws.SetLayout(uint(i))
		}
	}
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func HorizontalTopLayout(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "horizontal-top" {
			ws.SetLayout(uint(i))
		}
	}
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func HorizontalBottomLayout(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "horizontal-bottom" {
			ws.SetLayout(uint(i))
		}
	}
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func MaximizedLayout(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "maximized" {
			ws.SetLayout(uint(i))
		}
	}
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func FullscreenLayout(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "fullscreen" {
			ws.SetLayout(uint(i))
		}
	}
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func MakeMaster(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	if c, ok := tr.Clients[store.Windows.Active.Id]; ok {
		ws.ActiveLayout().MakeMaster(c)
		tr.Tile(ws)
		return true
	}

	return false
}

func MakeMasterNext(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	c := ws.ActiveLayout().NextClient()
	if c == nil {
		return false
	}

	ws.ActiveLayout().MakeMaster(c)
	tr.Tile(ws)

	return NextWindow(tr, ws)
}

func MakeMasterPrevious(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	c := ws.ActiveLayout().PreviousClient()
	if c == nil {
		return false
	}

	ws.ActiveLayout().MakeMaster(c)
	tr.Tile(ws)

	return PreviousWindow(tr, ws)
}

func IncreaseMaster(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	ws.ActiveLayout().IncreaseMaster()
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func DecreaseMaster(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	ws.ActiveLayout().DecreaseMaster()
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func IncreaseSlave(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	ws.ActiveLayout().IncreaseSlave()
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func DecreaseSlave(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	ws.ActiveLayout().DecreaseSlave()
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func IncreaseProportion(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	ws.ActiveLayout().IncreaseProportion()
	tr.Tile(ws)

	return true
}

func DecreaseProportion(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	ws.ActiveLayout().DecreaseProportion()
	tr.Tile(ws)

	return true
}

func NextWindow(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	c := ws.ActiveLayout().NextClient()
	if c == nil {
		return false
	}

	store.ActiveWindowSet(store.X, c.Window)

	return true
}

func PreviousWindow(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	c := ws.ActiveLayout().PreviousClient()
	if c == nil {
		return false
	}

	store.ActiveWindowSet(store.X, c.Window)

	return true
}

func Reset(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.TilingDisabled() {
		return false
	}
	ws.ResetLayouts()
	tr.Tile(ws)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func Exit(tr *desktop.Tracker) bool {
	tr.Write()

	xevent.Detach(store.X, store.X.RootWin())

	for _, ws := range tr.Workspaces {
		if ws.TilingDisabled() {
			continue
		}
		ws.DisableTiling()
		tr.Restore(ws, store.Latest)
	}

	log.Info("Exit")

	os.Exit(0)

	return true
}

func External(command string) bool {
	params := strings.Split(command, " ")

	if !common.HasFlag("enable-external-commands") {
		log.Warn("Executing external command \"", params[0], "\" disabled")
		return false
	}

	log.Info("Executing external command \"", params[0], " ", params[1:], "\"")

	// Execute external command
	cmd := exec.Command(params[0], params[1:]...)
	if err := cmd.Run(); err != nil {
		log.Error(err)
		return false
	}

	return true
}

func OnExecute(fun func(string, uint, uint)) {
	executeCallbacksFun = append(executeCallbacksFun, fun)
}

func executeCallbacks(action string, desk uint, screen uint) {
	log.Info("Execute event ", action)

	for _, fun := range executeCallbacksFun {
		fun(action, desk, screen)
	}
}

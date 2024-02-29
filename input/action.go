package input

import (
	"os"
	"os/exec"
	"strings"

	"golang.org/x/exp/maps"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/desktop"
	"github.com/leukipp/cortile/v2/store"
	"github.com/leukipp/cortile/v2/ui"

	log "github.com/sirupsen/logrus"
)

var (
	executeCallbacksFun []func(string) // Execute events callback functions
)

func Execute(action string, mod string, tr *desktop.Tracker) bool {
	success := false
	if len(strings.TrimSpace(action)) == 0 {
		return false
	}

	log.Info("Execute action [", action, "-", mod, "]")

	for _, ws := range tr.Workspaces {
		active := tr.ActiveWorkspace()

		// Execute only on active screen
		if mod == "current" && ws.Location != active.Location {
			continue
		}

		// Execute only on active workspace
		if mod == "screens" && (ws.Location.DeskNum != active.Location.DeskNum) {
			continue
		}

		// Choose action command
		switch action {
		case "enable":
			success = Enable(tr, ws)
		case "disable":
			success = Disable(tr, ws)
		case "restore":
			success = Restore(tr, ws)
		case "toggle":
			success = Toggle(tr, ws)
		case "cycle_next":
			success = CycleNext(tr, ws)
		case "cycle_previous":
			success = CyclePrevious(tr, ws)
		case "layout_fullscreen":
			success = FullscreenLayout(tr, ws)
		case "layout_vertical_left":
			success = VerticalLeftLayout(tr, ws)
		case "layout_vertical_right":
			success = VerticalRightLayout(tr, ws)
		case "layout_horizontal_top":
			success = HorizontalTopLayout(tr, ws)
		case "layout_horizontal_bottom":
			success = HorizontalBottomLayout(tr, ws)
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

		if !success {
			return false
		}

		// Notify socket
		type Action struct {
			DeskNum   uint
			ScreenNum uint
		}
		NotifySocket(Message[Action]{
			Type: "Action",
			Name: action,
			Data: Action{DeskNum: ws.Location.DeskNum, ScreenNum: ws.Location.ScreenNum},
		})
	}

	// Execute callbacks
	executeCallbacks(action)

	return true
}

func Query(state string, tr *desktop.Tracker) bool {
	success := false
	if len(strings.TrimSpace(state)) == 0 {
		return false
	}

	log.Info("Query state [", state, "]")

	ws := tr.ActiveWorkspace()

	// Choose state query
	switch state {
	case "workspaces":
		type Workspaces struct {
			DeskNum    uint
			ScreenNum  uint
			Workspaces []*desktop.Workspace
		}
		NotifySocket(Message[Workspaces]{
			Type: "State",
			Name: state,
			Data: Workspaces{DeskNum: ws.Location.DeskNum, ScreenNum: ws.Location.ScreenNum, Workspaces: maps.Values(tr.Workspaces)},
		})
		success = true
	case "arguments":
		NotifySocket(Message[common.Arguments]{
			Type: "State",
			Name: state,
			Data: common.Args,
		})
		success = true
	case "configs":
		NotifySocket(Message[common.Configuration]{
			Type: "State",
			Name: state,
			Data: common.Config,
		})
		success = true
	}

	return success
}

func Enable(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	ws.Enable()
	tr.Update()
	ws.Tile()

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func Disable(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	ws.Disable()
	ws.Restore(store.Latest)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func Restore(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	ws.Disable()
	ws.Restore(store.Original)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func Toggle(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return Enable(tr, ws)
	}
	return Disable(tr, ws)
}

func CycleNext(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	ws.CycleLayout(1)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func CyclePrevious(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	ws.CycleLayout(-1)

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func FullscreenLayout(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "fullscreen" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func VerticalLeftLayout(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "vertical-left" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func VerticalRightLayout(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "vertical-right" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func HorizontalTopLayout(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "horizontal-top" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func HorizontalBottomLayout(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "horizontal-bottom" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func MakeMaster(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	if c, ok := tr.Clients[store.ActiveWindow]; ok {
		ws.ActiveLayout().MakeMaster(c)
		ws.Tile()
		return true
	}

	return false
}

func MakeMasterNext(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	c := ws.ActiveLayout().NextClient()
	if c == nil {
		return false
	}

	ws.ActiveLayout().MakeMaster(c)
	ws.Tile()

	return NextWindow(tr, ws)
}

func MakeMasterPrevious(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	c := ws.ActiveLayout().PreviousClient()
	if c == nil {
		return false
	}

	ws.ActiveLayout().MakeMaster(c)
	ws.Tile()

	return PreviousWindow(tr, ws)
}

func IncreaseMaster(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	ws.ActiveLayout().IncreaseMaster()
	ws.Tile()

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func DecreaseMaster(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	ws.ActiveLayout().DecreaseMaster()
	ws.Tile()

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func IncreaseSlave(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	ws.ActiveLayout().IncreaseSlave()
	ws.Tile()

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func DecreaseSlave(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	ws.ActiveLayout().DecreaseSlave()
	ws.Tile()

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func IncreaseProportion(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	ws.ActiveLayout().IncreaseProportion()
	ws.Tile()

	return true
}

func DecreaseProportion(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	ws.ActiveLayout().DecreaseProportion()
	ws.Tile()

	return true
}

func NextWindow(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	c := ws.ActiveLayout().NextClient()
	if c == nil {
		return false
	}

	c.Activate()

	return true
}

func PreviousWindow(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	c := ws.ActiveLayout().PreviousClient()
	if c == nil {
		return false
	}

	c.Activate()

	return true
}

func Reset(tr *desktop.Tracker, ws *desktop.Workspace) bool {
	if ws.Disabled() {
		return false
	}
	ws.ResetLayouts()
	ws.Tile()

	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	return true
}

func Exit(tr *desktop.Tracker) bool {
	tr.Write()

	for _, ws := range tr.Workspaces {
		if ws.Disabled() {
			continue
		}
		ws.Disable()
		ws.Restore(store.Latest)
	}

	log.Info("Exit")

	os.Remove(common.Args.Sock + ".in")
	os.Remove(common.Args.Sock + ".out")

	os.Exit(1)

	return true
}

func External(command string) bool {
	params := strings.Split(command, " ")

	log.Info("Execute command ", params[0], " ", params[1:])

	// Execute external command
	cmd := exec.Command(params[0], params[1:]...)
	if err := cmd.Run(); err != nil {
		log.Error(err)
		return false
	}

	return true
}

func OnExecute(fun func(string)) {
	executeCallbacksFun = append(executeCallbacksFun, fun)
}

func executeCallbacks(arg string) {
	log.Info("Execute event ", arg)

	for _, fun := range executeCallbacksFun {
		fun(arg)
	}
}

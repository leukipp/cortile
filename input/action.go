package input

import (
	"os"
	"os/exec"
	"strings"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"

	log "github.com/sirupsen/logrus"
)

func Execute(a string, tr *desktop.Tracker) bool {
	success := false
	if len(strings.TrimSpace(a)) == 0 {
		return false
	}

	log.Info("Execute action [", a, "]")

	switch a {
	case "tile":
		success = Tile(tr)
	case "untile":
		success = UnTile(tr)
	case "layout_cycle":
		success = SwitchLayout(tr)
	case "layout_fullscreen":
		success = FullscreenLayout(tr)
	case "layout_vertical_left":
		success = VerticalLeftLayout(tr)
	case "layout_vertical_right":
		success = VerticalRightLayout(tr)
	case "layout_horizontal_top":
		success = HorizontalTopLayout(tr)
	case "layout_horizontal_bottom":
		success = HorizontalBottomLayout(tr)
	case "master_make":
		success = MakeMaster(tr)
	case "master_increase":
		success = IncreaseMaster(tr)
	case "master_decrease":
		success = DecreaseMaster(tr)
	case "slave_increase":
		success = IncreaseSlave(tr)
	case "slave_decrease":
		success = DecreaseSlave(tr)
	case "proportion_increase":
		success = IncreaseProportion(tr)
	case "proportion_decrease":
		success = DecreaseProportion(tr)
	case "window_next":
		success = NextWindow(tr)
	case "window_previous":
		success = PreviousWindow(tr)
	case "exit":
		success = Exit(tr)
	default:
		params := strings.Split(a, " ")
		log.Info("Execute command ", params[0], " ", params[1:])

		// Execute external command
		cmd := exec.Command(params[0], params[1:]...)
		if err := cmd.Run(); err != nil {
			log.Error(err)
			return false
		}
		success = true
	}

	// Notify socket
	if success {
		type Action struct {
			Desk   uint
			Screen uint
		}
		NotifySocket(Message[Action]{
			Type: "Action",
			Name: a,
			Data: Action{Desk: common.CurrentDesk, Screen: common.CurrentScreen},
		})
	}

	return success
}

func Query(s string, tr *desktop.Tracker) bool {
	success := false
	if len(strings.TrimSpace(s)) == 0 {
		return false
	}

	log.Info("Query state [", s, "]")

	switch s {
	case "workspaces":
		type Workspaces struct {
			Desk       uint
			Screen     uint
			Workspaces []*desktop.Workspace
		}
		ws := Workspaces{Desk: common.CurrentDesk, Screen: common.CurrentScreen}
		for _, v := range tr.Workspaces {
			ws.Workspaces = append(ws.Workspaces, v)
		}
		NotifySocket(Message[Workspaces]{
			Type: "State",
			Name: s,
			Data: ws,
		})
		success = true
	case "arguments":
		NotifySocket(Message[common.Arguments]{
			Type: "State",
			Name: s,
			Data: common.Args,
		})
		success = true
	case "configs":
		NotifySocket(Message[common.ConfigMapper]{
			Type: "State",
			Name: s,
			Data: common.Config,
		})
		success = true
	}

	return success
}

func Tile(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	ws.Enable(true)
	tr.Update(true)

	desktop.ShowLayout(ws)

	return true
}

func UnTile(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	ws.Enable(false)
	ws.UnTile()

	desktop.ShowLayout(ws)

	return true
}

func SwitchLayout(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	ws.SwitchLayout()

	desktop.ShowLayout(ws)

	return true
}

func FullscreenLayout(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "fullscreen" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	desktop.ShowLayout(ws)

	return true
}

func VerticalLeftLayout(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "vertical-left" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	desktop.ShowLayout(ws)

	return true
}

func VerticalRightLayout(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "vertical-right" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	desktop.ShowLayout(ws)

	return true
}

func HorizontalTopLayout(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "horizontal-top" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	desktop.ShowLayout(ws)

	return true
}

func HorizontalBottomLayout(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "horizontal-bottom" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	desktop.ShowLayout(ws)

	return true
}

func MakeMaster(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	if c, ok := tr.Clients[common.ActiveWindow]; ok {
		ws.ActiveLayout().MakeMaster(c)
		ws.Tile()
		return true
	}
	return false
}

func IncreaseMaster(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	ws.ActiveLayout().IncreaseMaster()
	ws.Tile()

	desktop.ShowLayout(ws)

	return true
}

func DecreaseMaster(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	ws.ActiveLayout().DecreaseMaster()
	ws.Tile()

	desktop.ShowLayout(ws)

	return true
}

func IncreaseSlave(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	ws.ActiveLayout().IncreaseSlave()
	ws.Tile()

	desktop.ShowLayout(ws)

	return true
}

func DecreaseSlave(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	ws.ActiveLayout().DecreaseSlave()
	ws.Tile()

	desktop.ShowLayout(ws)

	return true
}

func IncreaseProportion(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	ws.ActiveLayout().IncreaseProportion()
	ws.Tile()

	return true
}

func DecreaseProportion(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	ws.ActiveLayout().DecreaseProportion()
	ws.Tile()

	return true
}

func NextWindow(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	ws.ActiveLayout().NextClient()

	return true
}

func PreviousWindow(tr *desktop.Tracker) bool {
	ws := tr.ActiveWorkspace()
	if !ws.IsEnabled() {
		return false
	}
	ws.ActiveLayout().PreviousClient()

	return true
}

func Exit(tr *desktop.Tracker) bool {
	for _, ws := range tr.Workspaces {
		if !ws.IsEnabled() {
			continue
		}
		ws.Enable(false)
		ws.UnTile()
	}

	os.Remove(common.Args.Sock + ".in")
	os.Remove(common.Args.Sock + ".out")

	os.Exit(1)

	return true
}

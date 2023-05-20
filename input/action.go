package input

import (
	"os"
	"os/exec"
	"strings"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"

	log "github.com/sirupsen/logrus"
)

func Execute(a string, tr *desktop.Tracker) {
	if len(strings.TrimSpace(a)) == 0 {
		return
	}
	log.Info("Execute action [", a, "]")

	switch a {
	case "tile":
		Tile(tr)
	case "untile":
		UnTile(tr)
	case "layout_cycle":
		SwitchLayout(tr)
	case "layout_fullscreen":
		FullscreenLayout(tr)
	case "layout_vertical_left":
		VerticalLeftLayout(tr)
	case "layout_vertical_right":
		VerticalRightLayout(tr)
	case "layout_horizontal_top":
		HorizontalTopLayout(tr)
	case "layout_horizontal_bottom":
		HorizontalBottomLayout(tr)
	case "master_make":
		MakeMaster(tr)
	case "master_increase":
		IncreaseMaster(tr)
	case "master_decrease":
		DecreaseMaster(tr)
	case "slave_increase":
		IncreaseSlave(tr)
	case "slave_decrease":
		DecreaseSlave(tr)
	case "proportion_increase":
		IncreaseProportion(tr)
	case "proportion_decrease":
		DecreaseProportion(tr)
	case "window_next":
		NextWindow(tr)
	case "window_previous":
		PreviousWindow(tr)
	case "exit":
		Exit(tr)
	default:
		params := strings.Split(a, " ")
		log.Info("Execute command ", params[0], " ", params[1:])

		// execute process
		cmd := exec.Command(params[0], params[1:]...)
		if err := cmd.Run(); err != nil {
			log.Error(err)
		}
	}
}

func Tile(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	ws.Enable(true)
	tr.Update()

	desktop.ShowLayout(ws)
}

func UnTile(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	ws.Enable(false)
	ws.UnTile()
}

func SwitchLayout(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	ws.SwitchLayout()

	desktop.ShowLayout(ws)
}

func FullscreenLayout(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "fullscreen" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	desktop.ShowLayout(ws)
}

func VerticalLeftLayout(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "vertical-left" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	desktop.ShowLayout(ws)
}

func VerticalRightLayout(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "vertical-right" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	desktop.ShowLayout(ws)
}

func HorizontalTopLayout(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "horizontal-top" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	desktop.ShowLayout(ws)
}

func HorizontalBottomLayout(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	for i, l := range ws.Layouts {
		if l.GetName() == "horizontal-bottom" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()

	desktop.ShowLayout(ws)
}

func MakeMaster(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	ws.ActiveLayout().MakeMaster(tr.Clients[common.ActiveWin])
	ws.Tile()
}

func IncreaseMaster(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	ws.ActiveLayout().IncreaseMaster()
	ws.Tile()

	desktop.ShowLayout(ws)
}

func DecreaseMaster(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	ws.ActiveLayout().DecreaseMaster()
	ws.Tile()

	desktop.ShowLayout(ws)
}

func IncreaseSlave(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	ws.ActiveLayout().IncreaseSlave()
	ws.Tile()

	desktop.ShowLayout(ws)
}

func DecreaseSlave(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	ws.ActiveLayout().DecreaseSlave()
	ws.Tile()

	desktop.ShowLayout(ws)
}

func IncreaseProportion(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	ws.ActiveLayout().IncreaseProportion()
	ws.Tile()
}

func DecreaseProportion(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	ws.ActiveLayout().DecreaseProportion()
	ws.Tile()
}

func NextWindow(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	ws.ActiveLayout().NextClient()
}

func PreviousWindow(tr *desktop.Tracker) {
	ws := tr.Workspaces[common.CurrentDesk]
	if !ws.IsEnabled() {
		return
	}
	ws.ActiveLayout().PreviousClient()
}

func Exit(tr *desktop.Tracker) {
	for _, ws := range tr.Workspaces {
		if !ws.IsEnabled() {
			continue
		}
		ws.Enable(false)
		ws.UnTile()
	}
	os.Exit(1)
}

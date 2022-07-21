package input

import (
	"os/exec"
	"strings"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"

	log "github.com/sirupsen/logrus"
)

func Execute(a string, t *desktop.Tracker) {
	if len(strings.TrimSpace(a)) == 0 {
		return
	}
	log.Info("Execute action [", a, "]")

	switch a {
	case "tile":
		Tile(t)
	case "untile":
		UnTile(t)
	case "layout_cycle":
		SwitchLayout(t)
	case "layout_vertical":
		VerticalLayout(t)
	case "layout_horizontal":
		HorizontalLayout(t)
	case "layout_fullscreen":
		FullscreenLayout(t)
	case "master_make":
		MakeMaster(t)
	case "master_increase":
		IncreaseMaster(t)
	case "master_decrease":
		DecreaseMaster(t)
	case "proportion_increase":
		IncreaseProportion(t)
	case "proportion_decrease":
		DecreaseProportion(t)
	case "window_next":
		NextWindow(t)
	case "window_previous":
		PreviousWindow(t)
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

func Tile(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	ws.TilingEnabled = true
	ws.Tile()
}

func UnTile(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	ws.TilingEnabled = false
	ws.UnTile()
}

func SwitchLayout(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	if !ws.TilingEnabled {
		return
	}
	ws.SwitchLayout()
}

func VerticalLayout(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	if !ws.TilingEnabled {
		return
	}
	for i, l := range ws.Layouts {
		if l.GetType() == "vertical" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()
}

func HorizontalLayout(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	if !ws.TilingEnabled {
		return
	}
	for i, l := range ws.Layouts {
		if l.GetType() == "horizontal" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()
}

func FullscreenLayout(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	if !ws.TilingEnabled {
		return
	}
	for i, l := range ws.Layouts {
		if l.GetType() == "fullscreen" {
			ws.SetLayout(uint(i))
		}
	}
	ws.Tile()
}

func MakeMaster(t *desktop.Tracker) {
	c := t.Clients[common.ActiveWin]
	ws := t.Workspaces[common.CurrentDesk]
	if !ws.TilingEnabled {
		return
	}
	ws.ActiveLayout().MakeMaster(c)
	ws.Tile()
}

func IncreaseMaster(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	if !ws.TilingEnabled {
		return
	}
	ws.ActiveLayout().IncreaseMaster()
	ws.Tile()
}

func DecreaseMaster(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	if !ws.TilingEnabled {
		return
	}
	ws.ActiveLayout().DecreaseMaster()
	ws.Tile()
}

func IncreaseProportion(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	if !ws.TilingEnabled {
		return
	}
	ws.ActiveLayout().IncreaseProportion()
	ws.Tile()
}

func DecreaseProportion(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	if !ws.TilingEnabled {
		return
	}
	ws.ActiveLayout().DecreaseProportion()
	ws.Tile()
}

func NextWindow(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	if !ws.TilingEnabled {
		return
	}
	ws.ActiveLayout().NextClient()
}

func PreviousWindow(t *desktop.Tracker) {
	ws := t.Workspaces[common.CurrentDesk]
	if !ws.TilingEnabled {
		return
	}
	ws.ActiveLayout().PreviousClient()
}

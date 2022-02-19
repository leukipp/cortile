package main

import (
	"flag"
	"io"
	"os"

	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"
	"github.com/leukipp/cortile/input"

	log "github.com/sirupsen/logrus"
)

func main() {

	// TODO: allow only one instance

	// init log
	setLogLevel()

	// init state
	common.Init()

	// init workspace and tracker
	workspaces := desktop.CreateWorkspaces()
	tracker := desktop.CreateTracker(workspaces)

	// auto tile on startup
	if common.Config.StartupTiling {
		for _, ws := range workspaces {
			ws.Tile()
		}
	}

	// bind keys and mouse
	input.BindKeys(tracker)
	input.BindMouse(tracker)

	// run X event loop
	xevent.Main(common.X)
}

func setLogLevel() {
	var l string
	var vvv bool
	var vv bool
	var v bool

	flag.StringVar(&l, "l", "/tmp/cortile.log", "path of the logfile")
	flag.BoolVar(&vvv, "vvv", false, "very very verbose mode")
	flag.BoolVar(&vv, "vv", false, "very verbose mode")
	flag.BoolVar(&v, "v", false, "verbose mode")
	flag.Parse()

	if vvv {
		log.SetLevel(log.TraceLevel)
	} else if vv {
		log.SetLevel(log.DebugLevel)
	} else if v {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})

	file, err := os.OpenFile(l, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if file == nil {
		log.Error(err)
		return
	}

	log.SetOutput(io.MultiWriter(os.Stderr, file))
	log.RegisterExitHandler(func() {
		if file != nil {
			file.Close()
		}
	})
}

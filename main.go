package main

import (
	"flag"

	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/blrsn/zentile/state"
	log "github.com/sirupsen/logrus"
)

func main() {
	setLogLevel()
	state.Populate()

	t := initTracker(CreateWorkspaces())
	bindKeys(t)

	// Run X event loop
	xevent.Main(state.X)
}

func setLogLevel() {
	var verbose bool
	flag.BoolVar(&verbose, "v", false, "verbose mode")
	flag.Parse()

	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}

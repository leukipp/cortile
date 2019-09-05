package main

import (
	"flag"

	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xevent"
	"github.com/blrsn/zentile/state"
	log "github.com/sirupsen/logrus"
)

func main() {
	setLogLevel()
	checkEwmhCompliance()

	t := initTracker(CreateWorkspaces())
	bindKeys(t)

	// Run X event loop
	xevent.Main(state.X)
}

func checkEwmhCompliance() {
	_, err := ewmh.GetEwmhWM(state.X)
	if err != nil {
		log.Fatal("Window manager is not EWMH complaint!")
	}
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

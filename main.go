package main

import (
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"syscall"

	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"
	"github.com/leukipp/cortile/input"

	log "github.com/sirupsen/logrus"
)

//go:embed config.toml
var defaultConfig []byte

func main() {

	// Allow only one instance
	lock, err := createLockFile("/run/lock/cortile.lock")
	if err != nil {
		fmt.Println(fmt.Errorf("cortile already running (%s)", err))
		return
	}
	defer lock.Close()

	// Init log
	setLogLevel()

	// Init config
	common.InitConfig(defaultConfig)

	// Init state
	common.InitState()

	// Init workspace and tracker
	workspaces := desktop.CreateWorkspaces()
	tracker := desktop.CreateTracker(workspaces)

	// Tile on startup
	if common.Config.TilingEnabled {
		for _, ws := range workspaces {
			ws.Tile()
		}
	}

	// Bind keys and mouse
	input.BindKeys(tracker)
	input.BindMouse(tracker)

	// Run X event loop
	xevent.Main(common.X)
}

func createLockFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(fmt.Errorf("lock error (%s)", err))
		return nil, nil
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}

func setLogLevel() {
	var l string
	var vvv bool
	var vv bool
	var v bool

	flag.StringVar(&l, "l", "/tmp/cortile.log", "log path")
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

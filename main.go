package main

import (
	_ "embed"
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

var (
	version = "dev"     // Build version
	commit  = "local"   // Build commit
	date    = "unknown" // Build date
)

func main() {

	// Init command line arguments
	common.InitArgs(version, commit, date)

	// Init lock and log files
	defer InitLock().Close()
	InitLog()

	// Init config and state
	common.InitConfig(defaultConfig)
	common.InitState()

	// Init workspace and tracker
	workspaces := desktop.CreateWorkspaces()
	tracker := desktop.CreateTracker(workspaces)

	// Bind input events
	input.BindSignal(tracker)
	input.BindSocket(tracker)
	input.BindMouse(tracker)
	input.BindKeys(tracker)

	// Run X event loop
	xevent.Main(common.X)
}

func InitLock() *os.File {
	file, err := createLockFile(common.Args.Lock)
	if err != nil {
		fmt.Println(fmt.Errorf("cortile already running (%s)", err))
		os.Exit(1)
	}
	return file
}

func InitLog() *os.File {
	if common.Args.VVV {
		log.SetLevel(log.TraceLevel)
	} else if common.Args.VV {
		log.SetLevel(log.DebugLevel)
	} else if common.Args.V {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})

	file, err := createLogFile(common.Args.Log)
	if err != nil {
		return file
	}

	log.SetOutput(io.MultiWriter(os.Stderr, file))
	log.RegisterExitHandler(func() {
		if file != nil {
			file.Close()
		}
	})

	return file
}

func createLockFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(fmt.Errorf("FILE error (%s)", err))
		return nil, nil
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}

func createLogFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(fmt.Errorf("FILE error (%s)", err))
		return nil, err
	}
	return file, nil
}

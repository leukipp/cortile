package main

import (
	_ "embed"

	"fmt"
	"io"
	"os"
	"syscall"

	"runtime/debug"

	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"
	"github.com/leukipp/cortile/input"
	"github.com/leukipp/cortile/store"
	"github.com/leukipp/cortile/ui"

	log "github.com/sirupsen/logrus"
)

var (
	// Build name
	name = "cortile"

	// Build version
	version = "0.0.0"

	// Build commit
	commit = "local"

	// Build date
	date = "unknown"

	// Build source
	source = "https://github.com/leukipp/cortile"
)

var (
	//go:embed config.toml
	toml []byte

	//go:embed assets/images/logo.png
	icon []byte
)

func main() {

	// Init command line arguments
	common.InitArgs(name, version, commit, date, source)

	// Init embedded files
	common.InitFiles(toml, icon)

	// Init lock and log files
	defer InitLock().Close()
	InitLog()

	// Init cache and config
	common.InitCache()
	common.InitConfig()

	// Init root window properties
	store.InitRoot()

	// Run main application
	run()
}

func run() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(fmt.Errorf("%s\n%s", err, debug.Stack()))
		}
	}()

	// Create workspaces and tracker
	workspaces := desktop.CreateWorkspaces()
	tracker := desktop.CreateTracker(workspaces)

	// Show initial layout
	ws := tracker.ActiveWorkspace()
	ui.ShowLayout(ws)
	ui.UpdateIcon(ws)

	// Bind input events
	input.BindSignal(tracker)
	input.BindSocket(tracker)
	input.BindMouse(tracker)
	input.BindKeys(tracker)
	input.BindTray(tracker)

	// Run X event loop
	xevent.Main(store.X)
}

func InitLock() *os.File {
	file, err := createLockFile(common.Args.Lock)
	if err != nil {
		fmt.Println(fmt.Errorf("%s already running (%s)", common.Build.Name, err))
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

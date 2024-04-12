package main

import (
	_ "embed"

	"fmt"
	"io"
	"os"
	"syscall"

	"runtime/debug"

	"github.com/jezek/xgbutil/xevent"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/desktop"
	"github.com/leukipp/cortile/v2/input"
	"github.com/leukipp/cortile/v2/store"
	"github.com/leukipp/cortile/v2/ui"

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

	// Init process and build informations
	common.InitInfo(name, version, commit, date, source)

	// Init command line arguments
	common.InitArgs(input.Introspect())

	// Init embedded files
	common.InitFiles(toml, icon)

	// Run dbus instance
	runDbus()

	// Run main instance
	runMain()
}

func runDbus() {
	property := len(common.Args.Dbus.Property) > 0
	method := len(common.Args.Dbus.Method) > 0
	listen := common.Args.Dbus.Listen

	// Receive dbus property
	if property {
		input.Property(common.Args.Dbus.Property)
	}

	// Execute dbus method
	if method {
		input.Method(common.Args.Dbus.Method, common.Args.Dbus.Args)
	}

	// Listen to dbus events
	if listen {
		go input.Listen(common.Args.Dbus.Args)
		select {}
	}

	// Prevent main instance start
	if property || method || listen {
		os.Exit(1)
	}
}

func runMain() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal(fmt.Errorf("%s\n%s", err, debug.Stack()))
		}
	}()

	// Init lock and log files
	defer InitLock().Close()
	InitLog()

	// Init cache and config
	common.InitCache()
	common.InitConfig()

	// Init root properties
	store.InitRoot()

	// Create tracker
	tracker := desktop.CreateTracker()
	ws := tracker.ActiveWorkspace()
	if ws.Enabled() {
		ui.ShowLayout(ws)
	}

	// Bind input events
	input.BindSignal(tracker)
	input.BindSocket(tracker)
	input.BindMouse(tracker)
	input.BindKeys(tracker)
	input.BindTray(tracker)
	input.BindDbus(tracker)

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

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
	name string = "cortile"

	// Build target
	target string = "unknown"

	// Build version
	version string = "0.0.0"

	// Build commit
	commit string = "local"

	// Build date
	date string = "unknown"

	// Build source
	source string = "github.com/leukipp/cortile"

	// Build flags
	flags string
)

var (
	//go:embed config.toml
	toml []byte

	//go:embed assets/images/logo.png
	logo []byte
)

func main() {

	// Init process, build and source information
	common.InitInfo(name, target, version, commit, date, source, flags)

	// Init command line arguments
	common.InitArgs(input.Introspect())

	// Init embedded files
	common.InitFiles(toml, logo)

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
		input.Method(common.Args.Dbus.Method, common.Args.Dbus.P)
	}

	// Listen to dbus events
	if listen {
		go input.Listen(common.Args.Dbus.P)
		select {}
	}

	// Prevent main instance start
	if property || method || listen {
		os.Exit(0)
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

	// Create tracker instance
	tr := desktop.CreateTracker()
	input.Bind(tr)
	tr.Update()

	// Show layout overlay
	ws := tr.ActiveWorkspace()
	if ws.TilingEnabled() {
		ui.ShowLayout(ws)
	}

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

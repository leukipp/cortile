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

type Args struct {
	config string // Argument for config file path
	lock   string // Argument for lock file path
	log    string // Argument for log file path
	vvv    bool   // Argument for very very verbose mode
	vv     bool   // Argument for very verbose mode
	v      bool   // Argument for verbose mode
}

func main() {
	var args Args

	// Command line arguments
	flag.StringVar(&args.config, "config", common.ConfigFilePath(), "config file path")
	flag.StringVar(&args.lock, "lock", "/run/lock/cortile.lock", "lock file path")
	flag.StringVar(&args.log, "log", "/tmp/cortile.log", "log file path")
	flag.BoolVar(&args.vvv, "vvv", false, "very very verbose mode")
	flag.BoolVar(&args.vv, "vv", false, "very verbose mode")
	flag.BoolVar(&args.v, "v", false, "verbose mode")
	flag.Parse()

	// Allow only one instance
	lock, err := createLockFile(args.lock)
	if err != nil {
		fmt.Println(fmt.Errorf("cortile already running (%s)", err))
		return
	}
	defer lock.Close()

	// Init log
	setLogLevel(args)

	// Init config
	common.InitConfig(defaultConfig, args.config)

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
		desktop.ShowLayout(tracker.Workspaces[common.CurrentDesk])
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

func setLogLevel(args Args) {
	if args.vvv {
		log.SetLevel(log.TraceLevel)
	} else if args.vv {
		log.SetLevel(log.DebugLevel)
	} else if args.v {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
	log.SetFormatter(&log.TextFormatter{ForceColors: true, FullTimestamp: true})

	file, err := os.OpenFile(args.log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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

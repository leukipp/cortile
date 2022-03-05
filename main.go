package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"

	"github.com/BurntSushi/xgbutil/xevent"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"
	"github.com/leukipp/cortile/input"

	log "github.com/sirupsen/logrus"
)

func main() {

	// allow only one instance
	lock, err := createLockFile("/var/lock/cortile.lock")
	if err != nil {
		fmt.Println("cortile already running")
		return
	}
	defer lock.Close()

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

func createLockFile(filename string) (*os.File, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		file.Close()
		return nil, err
	}

	contents := strconv.Itoa(os.Getpid())
	if err := file.Truncate(0); err != nil {
		file.Close()
		return nil, err
	}

	if _, err := file.WriteString(contents); err != nil {
		file.Close()
		return nil, err
	}

	return file, nil
}

func setLogLevel() {
	var logfile string
	var vvv bool
	var vv bool
	var v bool

	flag.StringVar(&logfile, "logfile", "/tmp/cortile.log", "logfile path")
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

	file, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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

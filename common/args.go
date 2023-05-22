package common

import (
	_ "embed"
	"flag"
	"fmt"
)

var Args Arguments

type Arguments struct {
	Config string // Argument for config file path
	Lock   string // Argument for lock file path
	Sock   string // Argument for sock file path
	Log    string // Argument for log file path
	VVV    bool   // Argument for very very verbose mode
	VV     bool   // Argument for very verbose mode
	V      bool   // Argument for verbose mode
}

func InitArgs(version, commit, date string) {

	// Command line arguments
	flag.StringVar(&Args.Config, "config", ConfigFilePath(), "config file path")
	flag.StringVar(&Args.Lock, "lock", "/tmp/cortile.lock", "lock file path")
	flag.StringVar(&Args.Sock, "sock", "/tmp/cortile.sock", "sock file path")
	flag.StringVar(&Args.Log, "log", "/tmp/cortile.log", "log file path")
	flag.BoolVar(&Args.VVV, "vvv", false, "very very verbose mode")
	flag.BoolVar(&Args.VV, "vv", false, "very verbose mode")
	flag.BoolVar(&Args.V, "v", false, "verbose mode")

	// Command line usage text
	flag.CommandLine.Usage = func() {
		title := fmt.Sprintf("cortile v%s, built on %s (%s)", version, date, commit)
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\nUsage:\n", title)
		flag.PrintDefaults()
	}

	// Parse arguments
	flag.Parse()
}

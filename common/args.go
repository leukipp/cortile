package common

import (
	"flag"
	"fmt"
	"os"

	"path/filepath"
)

var (
	Args Arguments // Parsed arguments
)

type Arguments struct {
	Cache  string // Argument for cache folder path
	Config string // Argument for config file path
	Lock   string // Argument for lock file path
	Sock   string // Argument for sock file path
	Log    string // Argument for log file path
	VVV    bool   // Argument for very very verbose mode
	VV     bool   // Argument for very verbose mode
	V      bool   // Argument for verbose mode
	Dbus   struct {
		Listen   bool     // Argument for dbus listen flag
		Method   string   // Argument for dbus method name
		Property string   // Argument for dbus property name
		Args     []string // Argument for dbus method arguments
	}
}

func InitArgs(introspect map[string][]string) {

	// Command line arguments
	flag.StringVar(&Args.Cache, "cache", filepath.Join(CacheFolderPath(Build.Name), Build.Version), "cache folder path")
	flag.StringVar(&Args.Config, "config", filepath.Join(ConfigFolderPath(Build.Name), "config.toml"), "config file path")
	flag.StringVar(&Args.Lock, "lock", filepath.Join(os.TempDir(), fmt.Sprintf("%s.lock", Build.Name)), "lock file path")
	flag.StringVar(&Args.Sock, "sock", filepath.Join(os.TempDir(), fmt.Sprintf("%s.sock", Build.Name)), "sock file path (deprecated)")
	flag.StringVar(&Args.Log, "log", filepath.Join(os.TempDir(), fmt.Sprintf("%s.log", Build.Name)), "log file path")
	flag.BoolVar(&Args.VVV, "vvv", false, "very very verbose mode")
	flag.BoolVar(&Args.VV, "vv", false, "very verbose mode")
	flag.BoolVar(&Args.V, "v", false, "verbose mode")

	// Command line usage text
	flag.CommandLine.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\nUsage:\n", Build.Summary)
		flag.PrintDefaults()
	}
	flag.Parse()

	// Subcommand line arguments
	dbus := flag.NewFlagSet("dbus", flag.ExitOnError)
	dbus.BoolVar(&Args.Dbus.Listen, "listen", false, "dbus listen mode")
	dbus.StringVar(&Args.Dbus.Method, "method", "", "dbus method caller")
	dbus.StringVar(&Args.Dbus.Property, "property", "", "dbus property reader")
	Args.Dbus.Args = []string{}

	// Subcommand line usage text
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "dbus":
			dbus.Usage = func() {
				fmt.Fprintf(dbus.Output(), "%s\n\nUsage:\n", Build.Summary)
				dbus.PrintDefaults()

				if len(introspect) > 0 {
					if methods, ok := introspect["Methods"]; ok {
						fmt.Fprintf(dbus.Output(), "\nMethods:\n")
						for _, method := range methods {
							fmt.Fprintf(dbus.Output(), "  %s dbus -method %s\n", Build.Name, method)
						}
					}
					if properties, ok := introspect["Properties"]; ok {
						fmt.Fprintf(dbus.Output(), "\nProperties:\n")
						for _, property := range properties {
							fmt.Fprintf(dbus.Output(), "  %s dbus -property %s\n", Build.Name, property)
						}
					}
				} else {
					fmt.Fprintf(dbus.Output(), "\n>>> start %s to see further information's <<<\n", Build.Name)
				}
			}
			Args.Dbus.Args = ParseArgs(dbus, os.Args[2:])
		}

		// Check number of arguments
		if flag.NArg() == 1 {
			dbus.Usage()
			os.Exit(1)
		}
	}
}

func ParseArgs(flags *flag.FlagSet, args []string) []string {
	pargs := []string{}

	for {
		// Parse named arguments
		flags.Parse(args)

		// Check named arguments
		args = args[len(args)-flags.NArg():]
		if len(args) == 0 {
			break
		}

		// Check positional arguments
		pargs = append(pargs, args[0])
		args = args[1:]
	}

	// Parse positional arguments
	flags.Parse(pargs)

	return flags.Args()
}

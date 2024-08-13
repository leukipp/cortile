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
	Cache  string   // Argument for cache folder path
	Config string   // Argument for config file path
	Lock   string   // Argument for lock file path
	Log    string   // Argument for log file path
	VVV    bool     // Argument for very very verbose mode
	VV     bool     // Argument for very verbose mode
	V      bool     // Argument for verbose mode
	P      []string // Argument for positional values
	Dbus   struct {
		Listen   bool     // Argument for dbus listen flag
		Method   string   // Argument for dbus method name
		Property string   // Argument for dbus property name
		P        []string // Argument for dbus positional values
	}
}

func InitArgs(introspect map[string][]string) {

	// Command line arguments
	flag.StringVar(&Args.Cache, "cache", filepath.Join(CacheFolderPath(Build.Name), Build.Version), "cache folder path")
	flag.StringVar(&Args.Config, "config", filepath.Join(ConfigFolderPath(Build.Name), "config.toml"), "config file path")
	flag.StringVar(&Args.Lock, "lock", filepath.Join(os.TempDir(), fmt.Sprintf("%s.lock", Build.Name)), "lock file path")
	flag.StringVar(&Args.Log, "log", filepath.Join(os.TempDir(), fmt.Sprintf("%s.log", Build.Name)), "log file path")
	flag.BoolVar(&Args.VVV, "vvv", false, "very very verbose mode")
	flag.BoolVar(&Args.VV, "vv", false, "very verbose mode")
	flag.BoolVar(&Args.V, "v", false, "verbose mode")
	Args.P = []string{}

	// Command line usage text
	flag.CommandLine.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n\nUsage:\n", Build.Summary)
		flag.PrintDefaults()
	}

	// Parse command line arguments
	flag.Parse()
	Args.P = flag.Args()

	// Subcommand line arguments
	dbus := flag.NewFlagSet("dbus", flag.ExitOnError)
	dbus.BoolVar(&Args.Dbus.Listen, "listen", false, "dbus listen mode")
	dbus.StringVar(&Args.Dbus.Method, "method", "", "dbus method caller")
	dbus.StringVar(&Args.Dbus.Property, "property", "", "dbus property reader")
	Args.Dbus.P = []string{}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "dbus":

			// Subcommand line usage text
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

			// Parse subcommand line arguments
			FlagParse(dbus, os.Args[2:])
			Args.Dbus.P = dbus.Args()

			// Check subcommand line arguments
			if !Args.Dbus.Listen && len(Args.Dbus.Method) == 0 && len(Args.Dbus.Property) == 0 {
				dbus.Usage()
				os.Exit(2)
			}
		}
	}
}

func FlagParse(flags *flag.FlagSet, args []string) {
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
}

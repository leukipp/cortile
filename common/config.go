package common

import (
	"fmt"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"

	log "github.com/sirupsen/logrus"
)

var (
	Config Configuration // Decoded config values
)

type Configuration struct {
	TilingEnabled    bool              `toml:"tiling_enabled"`     // Tile windows on startup
	TilingLayout     string            `toml:"tiling_layout"`      // Initial tiling layout
	TilingGui        int               `toml:"tiling_gui"`         // Time duration of gui
	TilingIcon       [][]string        `toml:"tiling_icon"`        // Menu entries of systray
	WindowIgnore     [][]string        `toml:"window_ignore"`      // Regex to ignore windows
	WindowMastersMax int               `toml:"window_masters_max"` // Maximum number of allowed masters
	WindowSlavesMax  int               `toml:"window_slaves_max"`  // Maximum number of allowed slaves
	WindowGapSize    int               `toml:"window_gap_size"`    // Gap size between windows
	WindowDecoration bool              `toml:"window_decoration"`  // Show window decorations
	Proportion       float64           `toml:"proportion"`         // Master-slave area initial proportion
	ProportionStep   float64           `toml:"proportion_step"`    // Master-slave area step size proportion
	ProportionMin    float64           `toml:"proportion_min"`     // Window size minimum proportion
	EdgeMargin       []int             `toml:"edge_margin"`        // Margin values of tiling area
	EdgeCornerSize   int               `toml:"edge_corner_size"`   // Size of square defining edge corners
	EdgeCenterSize   int               `toml:"edge_center_size"`   // Length of rectangle defining edge centers
	Colors           map[string][]int  `toml:"colors"`             // List of color values for gui elements
	Keys             map[string]string `toml:"keys"`               // Event bindings for keyboard shortcuts
	Corners          map[string]string `toml:"corners"`            // Event bindings for hot-corner actions
	Systray          map[string]string `toml:"systray"`            // Event bindings for systray icon
}

func InitConfig() {

	// Create config folder if not exists
	configFolderPath := filepath.Dir(Args.Config)
	if _, err := os.Stat(configFolderPath); os.IsNotExist(err) {
		os.MkdirAll(configFolderPath, 0700)
	}

	// Write default config if not exists
	if _, err := os.Stat(Args.Config); os.IsNotExist(err) {
		ioutil.WriteFile(Args.Config, File.Toml, 0644)
	}

	// Read config file into memory
	readConfig(Args.Config)

	// Config file watcher
	watchConfig(Args.Config)
}

func ConfigFilePath(name string) string {

	// Obtain user home directory
	userHome, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error obtaining home directory ", err)
	}
	configFolderPath := filepath.Join(userHome, ".config", name)

	// Obtain config directory
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		configFolderPath = filepath.Join(xdgConfigHome, name)
	}

	return filepath.Join(configFolderPath, "config.toml")
}

func readConfig(configFilePath string) {
	fmt.Println(fmt.Errorf("LOAD %s [%s]", configFilePath, Build.Summary))
	log.Info("Starting [", Build.Summary, "]")

	// Decode contents into struct
	toml.DecodeFile(configFilePath, &Config)
}

func watchConfig(configFilePath string) {

	// Init file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err)
	} else {
		watcher.Add(configFilePath)
	}

	// Listen for events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Write) {
					readConfig(configFilePath)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error(err)
			}
		}
	}()
}

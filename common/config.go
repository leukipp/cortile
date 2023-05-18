package common

import (
	_ "embed"
	"fmt"

	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"

	log "github.com/sirupsen/logrus"
)

var Config ConfigMapper

type ConfigMapper struct {
	TilingEnabled    bool              `toml:"tiling_enabled"`     // Tile windows on startup
	TilingLayout     string            `toml:"tiling_layout"`      // Initial tiling layout
	TilingGui        int               `toml:"tiling_gui"`         // Time duration of gui
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
	Corners          map[string]string `toml:"corners"`            // Event bindings for hot-corners
}

func InitConfig(defaultConfig []byte, configFilePath string) {

	// Create config folder if not exists
	configFolderPath := filepath.Dir(configFilePath)
	if _, err := os.Stat(configFolderPath); os.IsNotExist(err) {
		os.MkdirAll(configFolderPath, 0700)
	}

	// Write default config if not exists
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		ioutil.WriteFile(configFilePath, defaultConfig, 0644)
	}

	// Read config file into memory
	readConfig(configFilePath)

	// Config file watcher
	watchConfig(configFilePath)
}

func ConfigFilePath() string {

	// Obtain user home directory
	userHome, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error obtaining home directory ", err)
	}
	configFolderPath := filepath.Join(userHome, ".config", "cortile")

	// Obtain config directory
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome != "" {
		configFolderPath = filepath.Join(xdgConfigHome, "cortile")
	}

	return filepath.Join(configFolderPath, "config.toml")
}

func readConfig(configFilePath string) {
	fmt.Println("LOAD", configFilePath)

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

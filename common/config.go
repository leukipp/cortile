package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"

	"github.com/mitchellh/go-homedir"

	log "github.com/sirupsen/logrus"
)

var Config ConfigMapper

type ConfigMapper struct {
	TilingEnabled    bool              `toml:"tiling_enabled"`    // Tile windows on startup
	TilingLayout     string            `toml:"tiling_layout"`     // Tile windows on startup
	Proportion       float64           `toml:"proportion"`        // Master-slave area initial proportion
	ProportionMin    float64           `toml:"proportion_min"`    // Master-slave area minimum proportion
	ProportionMax    float64           `toml:"proportion_max"`    // Master-slave area maximum proportion
	ProportionStep   float64           `toml:"proportion_step"`   // Master-slave area step size proportion
	WindowGap        int               `toml:"window_gap"`        // Gap size between windows
	WindowDecoration bool              `toml:"window_decoration"` // Show window decorations
	WindowIgnore     [][]string        `toml:"window_ignore"`     // Regex to ignore windows
	Keys             map[string]string `toml:"keys"`              // Key bindings for shortcuts
}

func init() {
	writeDefaultConfig()
	toml.DecodeFile(configFilePath(), &Config)
}

func writeDefaultConfig() {
	// Create config folder
	if _, err := os.Stat(configFolderPath()); os.IsNotExist(err) {
		os.MkdirAll(configFolderPath(), 0700)
	}

	// Write default config
	if _, err := os.Stat(configFilePath()); os.IsNotExist(err) {
		defaultConfig, err := ioutil.ReadFile("config.toml")
		if err != nil {
			log.Error(err)
		}
		ioutil.WriteFile(configFilePath(), defaultConfig, 0644)
	}
}

func configFolderPath() string {
	var configFolder string

	switch runtime.GOOS {
	case "linux":
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome != "" {
			configFolder = filepath.Join(xdgConfigHome, "cortile")
		} else {
			configFolder, _ = homedir.Expand("~/.config/cortile/")
		}
	default:
		configFolder, _ = homedir.Expand("~/.cortile/")
	}

	return configFolder
}

func configFilePath() string {
	return filepath.Join(configFolderPath(), "config.toml")
}

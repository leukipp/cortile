package common

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/BurntSushi/toml"

	"github.com/mitchellh/go-homedir"
)

var Config ConfigMapper

type ConfigMapper struct {
	Keybindings     map[string]string
	WindowsToIgnore [][]string `toml:"ignore"`
	Gap             int
	Division        float64
	Proportion      float64
	HideDecor       bool `toml:"remove_decorations"`
	StartupTiling   bool `toml:"startup_tiling"`
}

func init() {
	writeDefaultConfig()
	toml.DecodeFile(configFilePath(), &Config)
}

func writeDefaultConfig() {
	if _, err := os.Stat(configFolderPath()); os.IsNotExist(err) {
		os.MkdirAll(configFolderPath(), 0700)
	}

	if _, err := os.Stat(configFilePath()); os.IsNotExist(err) {
		ioutil.WriteFile(configFilePath(), []byte(defaultConfig), 0644)
	}
}

func configFolderPath() string {
	var configFolder string
	switch runtime.GOOS {
	case "linux":
		xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfigHome != "" {
			configFolder = filepath.Join(xdgConfigHome, "Cortile")
		} else {
			configFolder, _ = homedir.Expand("~/.config/Cortile/")
		}
	default:
		configFolder, _ = homedir.Expand("~/.Cortile/")
	}

	return configFolder
}

func configFilePath() string {
	return filepath.Join(configFolderPath(), "config.toml")
}

var defaultConfig = `TODO`

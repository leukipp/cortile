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
		defaultConfig, err := ioutil.ReadFile("config.toml")
		if err != nil {
			log.Error(err)
		}
		ioutil.WriteFile(configFilePath(), defaultConfig, 0644)
	}
}

func configFilePath() string {
	return filepath.Join(configFolderPath(), "config.toml")
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

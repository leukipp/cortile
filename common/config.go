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
		ioutil.WriteFile(configFilePath(), []byte(defaultConfig), 0644)
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

var defaultConfig = `# Tiling will be enabled on application start if set to true.
tiling_enabled = true

# Initial tiling layout ('vertical', 'horizontal', 'fullscreen')
tiling_layout = "vertical"

# Initial division of master-slave area.
proportion = 0.6

# Minimum division of master-slave area.
proportion_min = 0.1

# Maximum division of master-slave area.
proportion_max = 0.9

# How much to increment/decrement master-slave area.
proportion_step = 0.05

# How much space should be left between windows.
window_gap = 4

# Window decorations will be removed if set to false.
window_decoration = true

# Perl regex to ignore windows (['WM_CLASS', 'WM_NAME'] = ['ignore all windows with this class', 'but allow those with this name']).
# The WM_CLASS string name can be found by running 'xprop WM_CLASS'.
window_ignore = [
    ['xf.*', ''],
    ['nm.*', ''],
    ['gcr.*', ''],
    ['polkit.*', ''],
    ['wrapper.*', ''],
    ['lightdm.*', ''],
    ['blueman.*', ''],
    ['pavucontrol.*', ''],
    ['engrampa.*', ''],
    ['firefox.*', '.*Mozilla Firefox'],
]

[keys]
# You can view which keys activate which modifier using the 'xmodmap' program.
# Key symbols can be found by pressing keys using the 'xev' program.

# Tile the current workspace.
tile = "Control-Shift-t"

# Untile the current workspace.
untile = "Control-Shift-u"

# Make the active window as master.
make_active_window_master = "Control-Shift-m"

# Increase the number of masters.
increase_master = "Control-Shift-i"

# Decrease the number of masters.
decrease_master = "Control-Shift-d"

# Cycles through the available layouts.
switch_layout = "Control-Shift-s"

# Moves focus to the next window.
next_window = "Control-Shift-n"

# Moves focus to the previous window.
previous_window = "Control-Shift-p"

# Increases the size of the master windows.
increment_master = "Control-bracketright"

# Decreases the size of the master windows.
decrement_master = "Control-bracketleft"`

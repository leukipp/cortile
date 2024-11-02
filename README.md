# Cortile
![build](https://img.shields.io/github/actions/workflow/status/leukipp/cortile/release.yaml?style=flat-square)
![date](https://img.shields.io/github/release-date/leukipp/cortile?style=flat-square)
![downloads](https://img.shields.io/github/downloads/leukipp/cortile/total?style=flat-square)
![os](https://img.shields.io/badge/os-%20linux%20|%20freebsd%20-blue?style=flat-square)
![platform](https://img.shields.io/badge/platform-%20amd64%20|%20arm64%20|%20armv6%20|%20386%20-teal?style=flat-square)

<a href="https://github.com/leukipp/cortile"><img src="https://raw.githubusercontent.com/leukipp/cortile/main/assets/images/logo.png" style="display:inline-block;width:95px;margin-right:10px;" align="left"/></a>

Linux auto tiling manager with hot corner support for Openbox, Fluxbox, IceWM, Xfwm, KWin, Marco, Muffin, Mutter and other [EWMH](https://en.wikipedia.org/wiki/Extended_Window_Manager_Hints#List_of_window_managers_that_support_Extended_Window_Manager_Hints) compliant window managers using the [X11](https://en.wikipedia.org/wiki/X_Window_System) window system.
Therefore, this project provides dynamic tiling for XFCE, LXDE, LXQt, KDE and GNOME (Mate, Deepin, Cinnamon, Budgie) based desktop environments.

Simply keep your current window manager and install **cortile on top** of it.
Once enabled, the tiling manager will handle _resizing_ and _positioning_ of _existing_ and _new_ windows.
<br clear="left"/>

## Features [![features](https://img.shields.io/github/stars/leukipp/cortile?style=flat-square)](#features-)
- [x] Workspace based tiling.
- [x] Auto detection of panels.
- [x] Toggle window decorations.
- [x] User interface for tiling mode.
- [x] Systray icon indicator and menu.
- [x] Custom addons via python bindings.
- [x] Keyboard, hot corner and systray bindings.
- [x] Vertical, horizontal, maximized and fullscreen mode.
- [x] Remember layout proportions.
- [x] Floating and sticky windows.
- [x] Drag & drop window swap.
- [x] Workplace aware layouts.
- [x] Multi monitor support.

Support for **keyboard and mouse** events sets cortile apart from other tiling solutions.
The _go_ implementation ensures a fast and responsive system, where _multiple layouts_, _keyboard shortcuts_, _drag & drop_ and _hot corner_ events simplify and speed up your daily work.

[![demo](https://raw.githubusercontent.com/leukipp/cortile/main/assets/images/demo.gif)](https://github.com/leukipp/cortile/blob/main/assets/images/demo.gif)

## Installation [![installation](https://img.shields.io/github/v/release/leukipp/cortile?style=flat-square)](#installation-)
Manually [download](https://github.com/leukipp/cortile/releases/latest) the latest binary file from [releases](https://github.com/leukipp/cortile/releases/latest) or use wget:
```bash
wget -qO- $(wget -qO- https://api.github.com/repos/leukipp/cortile/releases/latest | \
jq -r '.assets[] | select(.name | contains ("linux_amd64.tar.gz")) | .browser_download_url') | \
tar -xvz
```

Execute the binary file and cortile will automatically begin tiling windows until you choose to stop it:
```bash
./cortile
```
Another installation method can be found in the [development](#development-) section.
The latest official release is published on GitHub.
Versions distributed via package managers are community supported and may be outdated.

### Service
To enable auto tiling on startup, you can run cortile as a service after the graphical user interface has been loaded.
A template to run cortile as a [systemd](https://en.wikipedia.org/wiki/Systemd) service is provided in the [services](https://github.com/leukipp/cortile/tree/main/assets/services) folder.
You may have to adjust the filepath/symlink under `ExecStart` and enable the user service:
```bash
# copy systemd service file
cp cortile.service ~/.config/systemd/user/

# reload systemd configuration
systemctl --user daemon-reload

# enable systemd service
systemctl --user enable cortile.service

# start systemd service
systemctl --user start cortile.service
```

### Usage
The layouts are based on the master-slave concept, where one side of the screen is considered to be the master area and the other side is considered to be the slave area:
- `vertical-right:` split the screen vertically, master area on the right.
- `vertical-left:` split the screen vertically, master area on the left.
- `horizontal-top:` split the screen horizontally, master area on the top.
- `horizontal-bottom:` split the screen horizontally, master area on the bottom.
- `maximized:` single window that fills the entire tiling area.
- `fullscreen:` single window that fills the entire screen.

The number of windows per side and the occupied space can be changed dynamically.
Adjustments to window sizes are considered to be proportion changes of the underlying layout.

Windows placed on the master side are static and the layout will only change as long the space is not fully occupied.
Once the master area is full, the slave area is used, where the layout changes dynamically based on available space and configuration settings.

## Configuration [![configuration](https://img.shields.io/badge/file-%20config.toml%20-gold?style=flat-square)](#configuration-)
The configuration file is located at `~/.config/cortile/config.toml` (or `XDG_CONFIG_HOME`) and is created with default values during the first startup.
Additional information about individual entries can be found in the comments section of the [config.toml](https://github.com/leukipp/cortile/blob/main/config.toml) file.

[![config](https://raw.githubusercontent.com/leukipp/cortile/main/assets/images/config.gif)](https://github.com/leukipp/cortile/blob/main/assets/images/config.gif)

### Shortcuts
The default keyboard shortcuts are assigned as shown below.
If some of them are already in use by your system, update the default values in the `[keys]` section of the configuration file:
| Keys                                                    | Description                                   |
| ------------------------------------------------------- | --------------------------------------------- |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Home</kbd>        | Enable tiling on the current screen           |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>End</kbd>         | Disable tiling on the current screen          |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>T</kbd>           | Toggle between enable and disable             |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>D</kbd>           | Toggle window decoration on and off           |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>R</kbd>           | Disable tiling and restore windows            |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>BackSpace</kbd>   | Reset layouts to default proportions          |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Next</kbd>        | Cycle through next layouts                    |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Prior</kbd>       | Cycle through previous layouts                |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Left</kbd>        | Activate vertical-left layout                 |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Right</kbd>       | Activate vertical-right layout                |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Up</kbd>          | Activate horizontal-top layout                |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Down</kbd>        | Activate horizontal-bottom layout             |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Space</kbd>       | Activate maximized layout                     |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Return</kbd>      | Activate fullscreen layout                    |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Plus</kbd>        | Increase number of maximum slave windows      |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Minus</kbd>       | Decrease number of maximum slave windows      |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_Add</kbd>      | Increase number of master windows             |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_Subtract</kbd> | Decrease number of master windows             |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_2</kbd>        | Move focus to the next window                 |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_8</kbd>        | Move focus to the previous window             |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_9</kbd>        | Move the active window to the next screen     |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_7</kbd>        | Move the active window to the previous screen |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_5</kbd>        | Make the active window master                 |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_6</kbd>        | Make the next window master                   |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_4</kbd>        | Make the previous window master               |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_3</kbd>        | Increase proportion of master-slave area      |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_1</kbd>        | Decrease proportion of master-slave area      |

Hot corner events are defined under the `[corners]` section and are triggered when the pointer enters one of the target areas:
| Corners                            | Description                              |
| ---------------------------------- | ---------------------------------------- |
| <kbd>Top</kbd>-<kbd>Left</kbd>     | Focus previous window                    |
| <kbd>Top</kbd>-<kbd>Right</kbd>    | Make the active window master            |
| <kbd>Bottom</kbd>-<kbd>Right</kbd> | Increase proportion of master-slave area |
| <kbd>Bottom</kbd>-<kbd>Left</kbd>  | Decrease proportion of master-slave area |

Systray events are defined under the `[systray]` section and are triggered when the pointer keys are pressed while hovering the icon:
| Pointer                            | Description                              |
| ---------------------------------- | ---------------------------------------- |
| <kbd>Middle</kbd>-<kbd>Click</kbd> | Toggle between enable and disable        |
| <kbd>Scroll</kbd>-<kbd>Up</kbd>    | Cycle through previous layouts           |
| <kbd>Scroll</kbd>-<kbd>Down</kbd>  | Cycle through next layouts               |
| <kbd>Scroll</kbd>-<kbd>Right</kbd> | Increase proportion of master-slave area |
| <kbd>Scroll</kbd>-<kbd>Left</kbd>  | Decrease proportion of master-slave area |

Common pointer shortcuts used in some environments:
- Move window: <kbd>Alt</kbd>+<kbd>Left-Click</kbd>.
- Resize window: <kbd>Alt</kbd>+<kbd>Right-Click</kbd>.
- Maximize window: <kbd>Alt</kbd>+<kbd>Double-Click</kbd>.

## Addons [![addons](https://img.shields.io/badge/api-%20dbus%20|%20python%20-red?style=flat-square)](#addons-)
External processes may communicate with cortile by using [dbus](https://en.wikipedia.org/wiki/D-Bus) directly or via the [cortile-addons](https://github.com/leukipp/cortile-addons) python bindings.

### D-Bus
Running `cortile` starts a dbus server instance that makes internal properties and method calls available.
Since using dbus communication directly with an external process, bash script, etc. is possible, the development requires some knowledge of dbus and is quite messy.

Therefore, there is a built-in dbus client incorporated in the same cortile binary that can be started via `cortile dbus -...` as a secondary process.
This client instance communicates with the running server instance and allows to listen for events and to execute remote procedure calls.

The documentation of available properties and method calls can be found via `cortile dbus -help`.

### Python
Additional python bindings are available to further simplify communication with cortile and to build a community-based library of useful snippets and examples.

For simplicity, the python bindings just spawn another cortile instance via `cortile dbus -...` running in the background and wrapping all available interfaces in easy-to-use python methods.

Example scripts and detailed information's on how to get started can be found in the [cortile-addons](https://github.com/leukipp/cortile-addons) repository.

## Development [![development](https://img.shields.io/github/go-mod/go-version/leukipp/cortile?label=go&style=flat-square)](#development-)
You need [go >= 1.22](https://go.dev/dl/) to compile cortile.

<details><summary>Install - go</summary><div>

### Option 1: Install go via package manager
Use a package manager supported on your system:
```bash
# apt
sudo apt install golang

# yum
sudo yum install golang

# dnf
sudo dnf install golang

# pacman
sudo pacman -S go
```

### Option 2: Install go via archive download
Download a binary release suitable for your system:
```bash
cd /tmp/ && wget https://dl.google.com/go/go1.22.8.linux-amd64.tar.gz
sudo tar -xvf go1.22.8.linux-amd64.tar.gz
sudo mv -fi go /usr/local
```

Set required environment variables:
```bash
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile
echo "export GOPATH=~/.go" >> ~/.profile
source ~/.profile
```

</div></details>

Verify the installed go version:
```bash
go env | grep "GOPATH\|GOVERSION"
```

<details><summary>Install - cortile</summary><div>

### Option 1: Install cortile via remote source
Install directly from develop branch:
```bash
go install github.com/leukipp/cortile/v2@develop
```

### Option 2: Install cortile via local source
Clone source code from develop branch:
```bash
git clone https://github.com/leukipp/cortile.git -b develop
cd cortile
```

If necessary you can make local changes, then execute:
```bash
go install -ldflags="-X 'main.date=$(date --iso-8601=seconds)'"
```

</div></details>

Start cortile in verbose mode:
```bash
$GOPATH/bin/cortile -v
```

## Additional [![additional](https://img.shields.io/github/issues-pr-closed/leukipp/cortile?style=flat-square)](#additional-)
Special use cases:
- Use the `window_slaves_max` property to limit the number of windows.
  - e.g. with one active master and `window_slaves_max = 2`, all windows following the third window are stacked behind the two slaves.
- Use the `edge_margin` property to account for additional spaces.
  - e.g. for deskbar panels or conky infographics.
- Use `tiling_enabled = false` if you prefer to enable tiling only when needed.
  - e.g. or to mainly utilize the hot corner functionalities.
- Use [cortile-addons](https://github.com/leukipp/cortile-addons) if you need any other specific logic.
  This repository offers a range of extensions and enhancements specifically designed for cortile.

Security concerns:
- Since the [dbus api](https://github.com/leukipp/cortile/tree/develop?tab=readme-ov-file#addons-) exposes internal cortile properties to the outside, malicious code running on the same host could easily access them.
  However, the information cortile holds (e.g. about open windows) can also be accessed using other tools interfacing with the X11 window system.
  Therefore the decision was made that direct access to cortile provides greater flexibility for running custom logic without compromising security.
  - If you want to disable this feature run cortile with `cortile disable-dbus-interface`.
- Any scripts placed in the `~/.config/cortile/addons/` folder will be executed when the application starts.
  This provides the possibility to run custom [cortile-addons](https://github.com/leukipp/cortile/tree/develop?tab=readme-ov-file#addons-) scripts without worrying much about startup behavior and dependency issues.
  However, it also creates a potential security risk, as malicious code could place files in this folder to be executed by cortile.
  - If you want to disable this feature run cortile with `cortile disable-addons-folder`.
- Newly pinned issues appear as menu entries in a submenu within the systray.
  This feature requires a network request to the GitHub API.
  - If you want to disable this feature run cortile with `cortile disable-issue-info`.
- Cortile checks for new releases and provides the option for an in-place upgrade of the current binary.
  Similar to the GitHub issue information, this feature also requires a network request.
  - If you want to disable this feature run cortile with `cortile disable-release-info`.  
- The binary file runs perfectly fine with user permissions.
  - Do not run cortile as root!

## Issues [![issues](https://img.shields.io/github/issues-closed/leukipp/cortile?style=flat-square)](#issues-)
Cortile works best with Xfwm and Openbox window systems.
However, it`s still possible that you may encounter problems during usage.

Windows:
- It's recommended to disable all build-in window snapping features (e.g. snap to other windows, snap to screen borders).
- It's recommended to disable any logic that changes the window focus other than by clicking or opening a window (e.g. focus follow mouse, scroll wheel focus). 
- Automatic panel detection may not work under some window managers, use the `edge_margin` property to adjust for additional margins.
- Particularly in GNOME based desktop environments, window displacements or resizing issues may occur.
- Sticky windows may cause unwanted layout modifications during workspace changes.
- Toggling window decoration may cause unwanted layout modifications.

Systray:
- Adjust the bindings in the `[systray]` section, as some pointer events may not fire across different desktop environments.
- Window managers not supporting [StatusNotifierItem](https://freedesktop.org/wiki/Specifications/StatusNotifierItem) for displaying systray icons will need to install [snixembed](https://github.com/fyne-io/systray#linuxbsd).

Debugging:
- If you encounter problems start the process with `cortile -vv`, which provides additional debug outputs.
- A log file is created by default under `/tmp/cortile.log`.

## Credits [![credits](https://img.shields.io/github/contributors/leukipp/cortile?style=flat-square)](#credits-)
Based on [zentile](https://github.com/blrsn/zentile) ([Berin Larson](https://github.com/blrsn)) and [pytyle3](https://github.com/BurntSushi/pytyle3) ([Andrew Gallant](https://github.com/BurntSushi)).  
The main libraries used in this project are [xgbutil](https://github.com/jezek/xgbutil), [toml](https://github.com/BurntSushi/toml), [dbus](https://github.com/godbus/dbus), [systray](https://github.com/fyne-io/systray), [gopsutil](https://github.com/shirou/gopsutil), [fsnotify](https://github.com/fsnotify/fsnotify), [selfupdate](https://github.com/minio/selfupdate) and [logrus](https://github.com/sirupsen/logrus).

## License [![license](https://img.shields.io/github/license/leukipp/cortile?style=flat-square)](#license-)
[MIT](https://github.com/leukipp/cortile/blob/main/LICENSE)

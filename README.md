# Cortile
<a href="https://github.com/leukipp/cortile"><img src="https://raw.githubusercontent.com/leukipp/cortile/main/assets/logo.png" style="display:inline-block;width:75px;margin-right:10px;" align="left"/></a>
Tiling manager with hot corner support for Xfce and other [EWMH Compliant Window Managers](https://en.wikipedia.org/wiki/Extended_Window_Manager_Hints#List_of_window_managers_that_support_Extended_Window_Manager_Hints).
Simply keep your current window manager and **install cortile on top** of it.
Once enabled, the tiling manager will handle _resizing_ and _positioning_ of _existing_ and _new_ windows.
<br clear="left"/>

## Features [![github](https://img.shields.io/github/stars/leukipp/cortile)](#features-)
- [x] Workspace based tiling.
- [x] Keyboard and hot corner events.
- [x] Vertical, horizontal and fullscreen mode.
- [x] Persistent windows via "Always on Visible Workspace".
- [x] Floating windows via "Always on Top".
- [x] Resize of master-slave area.
- [x] Drag & drop window swap.
- [x] Auto detection of panels.
- [x] Selective tiling areas.
- [x] Multi monitor support.

Support for **keyboard and mouse navigation** sets cortile apart from other tiling solutions.
The _go_ implementation ensures a fast and responsive system, where _multiple layouts_, _keyboard shortcuts_, _drag & drop_ and _hot corner_ events simplify and speed up your daily work.

[![demo](https://raw.githubusercontent.com/leukipp/cortile/main/assets/demo.gif)](https://github.com/leukipp/cortile/blob/main/assets/demo.gif)

## Installation [![github](https://img.shields.io/github/v/release/leukipp/cortile)](#installation-)
Download the latest binary file from the [releases](https://github.com/leukipp/cortile/releases/latest):
```bash
# extract cortile from the tar.gz archive
tar -xvf cortile_*_linux_amd64.tar.gz

# execute the binary file
./cortile
```
Alternative installation methods can be found in the [development](#development-) section.

## Configuration [![github](https://img.shields.io/github/release-date/leukipp/cortile)](#configuration-)
The configuration file is located at `~/.config/cortile/config.toml` and is created with default values during the first startup.
Additional information about individual entries can be found in the comments section of the [config.toml](https://github.com/leukipp/cortile/blob/main/config.toml) file.

[![config](https://raw.githubusercontent.com/leukipp/cortile/main/assets/config.gif)](https://github.com/leukipp/cortile/blob/main/assets/config.gif)

### Shortcuts
The default keyboard shortcuts are assigned as shown below.
If some of them are already in use by your system, update the default values in the `[keys]` section of the configuration file:
| Keys                                                    | Description                              |
| ------------------------------------------------------- | ---------------------------------------- |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>T</kbd>           | Tile current workspace                   |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>U</kbd>           | Untile current workspace                 |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>C</kbd>           | Cycle through layouts                    |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Space</kbd>       | Activate fullscreen layout               |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Left</kbd>        | Activate vertical-left layout            |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Right</kbd>       | Activate vertical-right layout           |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Top</kbd>         | Activate horizontal-top layout           |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Bottom</kbd>      | Activate horizontal-bottom layout        |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>M</kbd>           | Make the active window master            |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Plus</kbd>        | Increase number of master windows        |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>Minus</kbd>       | Decrease number of master windows        |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_Add</kbd>      | Increase number of maximum slave windows |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_Subtract</kbd> | Decrease number of maximum slave windows |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_6</kbd>        | Increase proportion of master-slave area |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_4</kbd>        | Decrease proportion of master-slave area |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_2</kbd>        | Focus next window                        |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>KP_8</kbd>        | Focus previous window                    |

Hot corner events are defined under the `[corners]` section and will be triggered when the mouse enters one of the target areas:
| Corners                             | Description                              |
| ----------------------------------- | ---------------------------------------- |
| <kbd>Top</kbd>-<kbd>Left</kbd>      | Focus previous window                    |
| <kbd>Top</kbd>-<kbd>Center</kbd>    | Activate horizontal-top layout           |
| <kbd>Top</kbd>-<kbd>Right</kbd>     | Make the active window master            |
| <kbd>Center</kbd>-<kbd>Right</kbd>  | Activate vertical-right layout           |
| <kbd>Bottom</kbd>-<kbd>Right</kbd>  | Increase proportion of master-slave area |
| <kbd>Bottom</kbd>-<kbd>Center</kbd> | Activate horizontal-bottom layout        |
| <kbd>Bottom</kbd>-<kbd>Left</kbd>   | Decrease proportion of master-slave area |
| <kbd>Center</kbd>-<kbd>Left</kbd>   | Activate vertical-left layout            |

Useful mouse shortcuts on Xfce environments:
- Move window: <kbd>Alt</kbd>+<kbd>Left-Click</kbd>.
- Resize window: <kbd>Alt</kbd>+<kbd>Right-Click</kbd>.
- Maximize window: <kbd>Alt</kbd>+<kbd>Double-Click</kbd>.

## Development [![github](https://img.shields.io/github/go-mod/go-version/leukipp/cortile)](#development-)
You need [go >= 1.17](https://go.dev/dl/) to compile cortile.

<details><summary>Install - go</summary><div>

### Option 1: Install go via package manager:
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

### Option 2: Install go via archive download:
Download a binary release suitable for your system:
```bash
cd /tmp/ && wget https://dl.google.com/go/go1.17.linux-amd64.tar.gz
sudo tar -xvf go1.17.linux-amd64.tar.gz
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
Install directly from main branch:
```bash
go install github.com/leukipp/cortile@main
```

### Option 2: Install cortile via local source
Clone source code from main branch:
```bash
git clone https://github.com/leukipp/cortile.git -b main
cd cortile
```

If necessary you can make local changes, then execute:
```bash
go build && go install
```

</div></details>

Start cortile in verbose mode:
```bash
$GOPATH/bin/cortile -v
```

## Additional [![github](https://img.shields.io/github/issues-pr-closed/leukipp/cortile)](#additional-)
Special settings:
- Use the `edge_margin` property to account for additional spaces.
  - e.g. panels (deskbar), desktop infographics (conky).
- Use the `edge_margin` property to enable tiling only for parts of the monitor. 
  - e.g. use a left/right margin that is half the resolution of a large display to tile only windows that are moved inside a certain area.
- Use the `window_slaves_max` property to limit the number of windows.
  - e.g. with one active master and `window_slaves_max = 2` the fourth and the following windows are stacked (behind the two slaves).

Hot corners:
- Use `tiling_enabled = false` if you prefer to utilize only the hot corner functionalities.
- Use the hot `[corners]` properties to execute any external command available on your system.
  - e.g. use `bottom_center = "firefox"` to open a web browser window.

Companion tools:
- You can install a [minimal-gtk](https://www.xfce-look.org/p/1016504) theme and leave `window_decoration = true`.
- Simply add cortile to your startup applications to run it after login.

## Issues [![github](https://img.shields.io/github/issues-closed/leukipp/cortile)](#issues-)
It's recommended to disable all build-in window snapping functions. On Xfce environments they can be found under "Window Manager" > "Advanced" > "Window snapping".

If you encounter problems start the process with `cortile -vv`, which provides additional verbose outputs.
A log file is created by default under `/tmp/cortile.log`.

Known limitations:
- Only the biggest monitor is used for tiling.
- Persistent window resize only for the master-slave proportion.

## Credits [![github](https://img.shields.io/github/contributors/leukipp/cortile)](#credits-)
Based on [zentile](https://github.com/blrsn/zentile) from [Berin Larson](https://github.com/blrsn).

## License [![github](https://img.shields.io/github/license/leukipp/cortile)](#license-)
[MIT](https://github.com/leukipp/cortile/blob/main/LICENSE)

# Cortile [WIP]
Tiling manager with hot corner support for Xfce and other [EWMH Compliant Window Managers](https://en.m.wikipedia.org/wiki/Extended_Window_Manager_Hints). Simply stick to your current window manager and install [cortile](https://github.com/leukipp/cortile) on top. Once enabled, the tiling manager takes care of resizing and positioning of existing and new windows.

## Features
- Workspace based tiling.
- Keyboard and hot corner events.
- Vertical, horizontal and fullscreen mode.
- Resize of master / slave area.
- Drag and drop window swap.
- Auto detection of panels.
- Multi monitor support.

## Install
### Requirements
Install [go](https://go.dev/) 1.17 with `apt`:
```bash
sudo apt install golang
```

Install [go](https://go.dev/) 1.17 with `pacman`:
```bash
sudo pacman -S go
```

### Use remote source
Install [cortile](https://github.com/leukipp/cortile) from GitHub `main` branch:
```bash
go install github.com/leukipp/cortile@main
```

### Use local source
Clone [cortile](https://github.com/leukipp/cortile) from GitHub `main` branch:
```bash
git clone https://github.com/leukipp/cortile.git -b main
cd cortile
```

Once you have made local changes run:
```bash
go build && go install
```

## Usage
Start in verbose mode:
```bash
cortile -v
```

### Xfce
Useful shortcuts for Xfce environments:
- Move window: <kbd>Alt</kbd>+<kbd>Left-Click</kbd>.
- Resize window: <kbd>Alt</kbd>+<kbd>Right-Click</kbd>.
- Maximize window: <kbd>Alt</kbd>+<kbd>Double-Click</kbd>.

## Config
The config file is located at `~/.config/cortile/config.toml`.

| Corner events                      | Description                       |
| ---------------------------------- | --------------------------------- |
| <kbd>Top</kbd>-<kbd>Left</kbd>     | Cycle through layouts             |
| <kbd>Top</kbd>-<kbd>Right</kbd>    | Make the active window master     |
| <kbd>Bottom</kbd>-<kbd>Right</kbd> | Increase number of master windows |
| <kbd>Bottom</kbd>-<kbd>Left</kbd>  | Decrease number of master windows |

| Key events                                    | Description                       |
| --------------------------------------------- | --------------------------------- |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>t</kbd> | Tile current workspace            |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>u</kbd> | Untile current workspace          |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>m</kbd> | Make the active window master     |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>n</kbd> | Goto next window                  |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>p</kbd> | Goto previous window              |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>s</kbd> | Cycle through layouts             |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>i</kbd> | Increase number of master windows |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>d</kbd> | Decrease number of master windows |
| <kbd>Ctrl</kbd>+<kbd>]</kbd>                  | Increment size of master windows  |
| <kbd>Ctrl</kbd>+<kbd>[</kbd>                  | Decrement size of master windows  |

## WIP
- Configurable hot corners (current: hardcoded corner events).
- Configurable LTR/RTL support (current: master is on the right side).
- Proper dual monitor support (current: only biggest monitor is tiled).
- Resizable windows (current: only master/slave proportion can be changed).

## Credits
Based on [zentile](https://github.com/blrsn/zentile) from [Berin Larson](https://github.com/blrsn/).

## License
[MIT](/LICENSE)

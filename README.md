# Cortile [WIP]
Tiling manager with hot corner support for Xfce, OpenBox and other [EWMH Compliant Window Managers](https://en.m.wikipedia.org/wiki/Extended_Window_Manager_Hints).

## Features
- Workspace based tiling.
- Keyboard and hot corner events.
- Vertical and horizontal tiling.
- Resize of master/slave area.
- Auto detection of panels.
- Multi monitor support.
- Customizable layouts.

## Install
### Remote source
Install from GitHub:
```bash
go get -u github.com/leukipp/cortile
go install github.com/leukipp/cortile
```

### Local source
Fetch from GitHub:
```bash
git clone https://github.com/leukipp/cortile.git
cd cortile
```

Make local changes and run:
```bash
go build
go install
```

## Run
Start in verbose mode:
```bash
cortile -v
```
Resizing of windows in Xfce can be done with <kbd>Alt</kbd>+<kbd>Right-Click</kbd>.

## Config
The config file is located at `~/.config/cortile/config.toml`.

| Corner events  | Description                       |
| -------------- | --------------------------------- |
| `top-left`     | Cycle through layouts             |
| `top-right`    | Make the active window as master  |
| `bottom-right` | Increase number of master windows |
| `bottom-left`  | Decrease number of master windows |

| Key events                                    | Description                       |
| --------------------------------------------- | --------------------------------- |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>t</kbd> | Tile current workspace            |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>u</kbd> | Untile current workspace          |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>m</kbd> | Make the active window as master  |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>i</kbd> | Increase number of master windows |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>d</kbd> | Decrease number of master windows |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>s</kbd> | Cycle through layouts             |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>n</kbd> | Goto next window                  |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>p</kbd> | Goto previous window              |
| <kbd>Ctrl</kbd>+<kbd>]</kbd>                  | Increase size of master windows   |
| <kbd>Ctrl</kbd>+<kbd>[</kbd>                  | Decrease size of master windows   |

## WIP
- Create default config (current: copy config template manually).
- Configurable hot corners (current: hardcoded corner events).
- Configurable LTR/RTL support (current: master is on the right side).
- Proper dual monitor support (current: only biggest monitor is tiled).
- Resizable windows (current: only master/slave proportion can be changed).

## Credits
Based on **zentile** from [Berin Larson](https://github.com/blrsn/):
- [https://github.com/blrsn/zentile](https://github.com/blrsn/zentile)

## License
[MIT](/LICENSE)

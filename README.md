# Cortile [WIP]
Tiling manager with hot corner support for Xfce and other [EWMH Compliant Window Managers](https://en.m.wikipedia.org/wiki/Extended_Window_Manager_Hints).
Simply keep your current window manager and install [cortile](https://github.com/leukipp/cortile) on top of it.
Once enabled, the tiling manager will handle resizing and positioning of existing and new windows.

## Features
- Workspace based tiling.
- Keyboard and hot corner events.
- Vertical, horizontal and fullscreen mode.
- Floating windows via "Always on Top".
- Resize of master / slave area.
- Drag and drop window swap.
- Auto detection of panels.
- Multi monitor support.

## Install
You need [go](https://go.dev/) >= 1.17 to run [cortile](https://github.com/leukipp/cortile).

### Requirements
Install go via `apt`:
```bash
sudo apt install golang
```

Install go via `pacman`:
```bash
sudo pacman -S go
```

Install go via `wget`:
```bash
cd /tmp && wget https://dl.google.com/go/go1.17.linux-amd64.tar.gz
sudo tar -xvf go1.17.linux-amd64.tar.gz
sudo mv -fi go /usr/local
```

```bash
echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.profile
echo "export GOPATH=~/.go" >> ~/.profile
source ~/.profile
```

### Cortile
Verify your go version >= 1.17:
```bash
go version
```

#### Use remote source
Install cortile from GitHub `main` branch:
```bash
go install github.com/leukipp/cortile@main
```

#### Use local source
Clone cortile from GitHub `main` branch:
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
~/.go/bin/cortile -v
```

In case of warnings during startup, check if [config.toml](https://github.com/leukipp/cortile/blob/main/config.toml) is properly configured.

### Xfce
Useful shortcuts for Xfce environments:
- Move window: <kbd>Alt</kbd>+<kbd>Left-Click</kbd>.
- Resize window: <kbd>Alt</kbd>+<kbd>Right-Click</kbd>.
- Maximize window: <kbd>Alt</kbd>+<kbd>Double-Click</kbd>.

## Config
The config file is located at `~/.config/cortile/config.toml`.

| Key events                                         | Description                       |
| -------------------------------------------------- | --------------------------------- |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>t</kbd>      | Tile current workspace            |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>u</kbd>      | Untile current workspace          |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>c</kbd>      | Cycle through layouts             |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>space</kbd>  | Activate fullscreen layout        |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>left</kbd>   | Activate vertical-left layout     |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>right</kbd>  | Activate vertical-right layout    |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>top</kbd>    | Activate horizontal-top layout    |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>bottom</kbd> | Activate horizontal-bottom layout |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>m</kbd>      | Make the active window master     |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>i</kbd>      | Increase number of master windows |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>d</kbd>      | Decrease number of master windows |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>plus</kbd>   | Increase size of master area      |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>minus</kbd>  | Decrease size of master area      |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>n</kbd>      | Focus next window                 |
| <kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>p</kbd>      | Focus previous window             |

| Corner events                       | Description                       |
| ----------------------------------- | --------------------------------- |
| <kbd>Top</kbd>-<kbd>Left</kbd>      | Cycle through layouts             |
| <kbd>Top</kbd>-<kbd>Center</kbd>    | Tile current workspace            |
| <kbd>Top</kbd>-<kbd>Right</kbd>     | Make the active window master     |
| <kbd>Center</kbd>-<kbd>Right</kbd>  | Increase size of master area      |
| <kbd>Bottom</kbd>-<kbd>Right</kbd>  | Increase number of master windows |
| <kbd>Bottom</kbd>-<kbd>Center</kbd> | Untile current workspace          |
| <kbd>Bottom</kbd>-<kbd>Left</kbd>   | Decrease number of master windows |
| <kbd>Center</kbd>-<kbd>Left</kbd>   | Decrease size of master area      |

## WIP
- Proper dual monitor support (current: only biggest monitor is tiled).
- Resizable windows (current: only master/slave proportion can be changed).
- Max windows count (current: more windows always subdivide slave area).

## Credits
Based on [zentile](https://github.com/blrsn/zentile) from [Berin Larson](https://github.com/blrsn/).

## License
[MIT](/LICENSE)

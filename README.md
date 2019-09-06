# zentile
Automatic Tiling for EWMH Complaint Window Managers.

![zentile screencast](docs/screencast.gif)

## FEATURES
- Zentile allows tiling on a per workspace basis. 
- Comes with two simple tiling layouts [Vertical & Horizontal]
- Customizable gap between tiling windows.
- Autodetection of panels and docks.
- Auto-hides window decorations on tiling.

## INSTALLATION
```
$ go get github.com/blrsn/zentile
$ "$GOPATH/bin/zentile"
```

[Binary releases](https://github.com/blrsn/zentile/releases) are also available.

### Commands

Keybinding                                          | Description
----------------------------------------------------|---------------------------------------
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>t</kbd>       | Tile current workspace 
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>u</kbd>       | Untile current workspace
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>s</kbd>       | Cycle through layouts
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>n</kbd>       | Goto next window
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>p</kbd>       | Goto previous window
<kbd>Ctrl</kbd>+<kbd>]</kbd>                        | Increase size of master windows
<kbd>Ctrl</kbd>+<kbd>[</kbd>                        | Decrease size of master windows
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>i</kbd>       | Increment number of master windows
<kbd>Ctrl</kbd>+<kbd>Shift</kbd>+<kbd>d</kbd>       | Decrement number of master windows

**Note:** zentile has been tested on Openbox.It should technically work with any ewmh complaint window manager.

### Credits

Inspired by BurntSushi's [pytyle](https://github.com/BurntSushi/pytyle3).  
Theme used in the screencast above, comes from addy-dclxvi's [openbox theme collection](https://github.com/addy-dclxvi/openbox-theme-collections).

## License

zentile is licensed under the MIT License. See the full license text in [`LICENSE`](LICENSE).

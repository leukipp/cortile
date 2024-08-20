package input

import (
	"fmt"
	"strings"
	"time"

	"os/exec"

	"fyne.io/systray"

	"github.com/godbus/dbus/v5"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/desktop"
	"github.com/leukipp/cortile/v2/store"
	"github.com/leukipp/cortile/v2/ui"

	log "github.com/sirupsen/logrus"
)

var (
	clicked bool          // Tray clicked state from dbus
	button  store.XButton // Pointer button state of device
	click   *time.Timer   // Timer to compress pointer events
	menu    *Menu         // Items collection of systray menu
)

type Menu struct {
	Toggle     *systray.MenuItem   // Toggle checkbox item
	Decoration *systray.MenuItem   // Decoration checkbox item
	Actions    []*systray.MenuItem // Actions for commands
}

func BindTray(tr *desktop.Tracker) {
	if len(common.Config.TilingIcon) == 0 {
		return
	}

	// Start systray icon
	go systray.Run(func() {
		items(tr)
		messages(tr)
	}, func() {})

	// Attach execute events
	OnExecute(func(action string, desktop uint, screen uint) {
		onExecute(tr, action)
	})

	// Attach pointer events
	store.OnPointerUpdate(func(pointer store.XPointer, desktop uint, screen uint) {
		onPointerClick(tr, pointer)
	})
}

func items(tr *desktop.Tracker) {
	systray.SetTooltip(fmt.Sprintf("%s - tiling manager", common.Build.Name))
	systray.SetTitle(common.Build.Name)

	// Version text
	title := fmt.Sprintf("%s v%s", common.Build.Name, common.Build.Version)
	version := systray.AddMenuItem(title, title)

	// Version icon
	version.SetIcon(common.File.Logo)
	if !common.HasReleaseInfos() && !common.HasIssueInfos() {
		version.Disable()
	}

	// Issue submenu
	if common.HasIssueInfos() {
		version.AddSubMenuItem("Issues", "Issues").Disable()
		for _, issue := range common.Source.Issues {
			title := fmt.Sprintf("%s #%d", issue.Name, issue.Id)
			subitem := version.AddSubMenuItem(title, title)

			// Issue hint icon
			subitem.SetIcon(ui.HintIcon(issue.Unseen()))

			// Issue item click
			go func(info common.Info) {
				for {
					<-subitem.ClickedCh

					// Open browser link
					exec.Command("xdg-open", info.Url).Start()

					// Update cache and ui icons
					if info.Seen() {
						subitem.SetIcon(ui.HintIcon(false))
						ui.UpdateIcon(tr.ActiveWorkspace())
					}
				}
			}(issue)
		}
	}

	// Separator submenu
	if common.HasReleaseInfos() && common.HasIssueInfos() {
		version.AddSeparator()
	}

	// Release submenu
	if common.HasReleaseInfos() {
		version.AddSubMenuItem("Releases", "Releases").Disable()
		for _, release := range common.Source.Releases {
			title := fmt.Sprintf("Release %s v%s is available", common.Build.Name, release.Name)
			subitem := version.AddSubMenuItem(title, title)

			// Release hint icon
			icon := ui.HintIcon(release.Unseen())
			subitem.SetIcon(icon)

			// Release item click
			go func(info common.Info) {
				for {
					<-subitem.ClickedCh

					// Open browser link
					exec.Command("xdg-open", info.Url).Start()

					// Update cache and ui icons
					if info.Seen() {
						subitem.SetIcon(ui.HintIcon(false))
						ui.UpdateIcon(tr.ActiveWorkspace())
					}
				}
			}(release)
		}
	}

	// Update submenu
	if common.HasReleaseInfos() {
		version.AddSubMenuItem("Updates", "Updates").Disable()
		for _, release := range common.Source.Releases {
			major, minor, patch := common.SemverUpdateInfos()
			title := fmt.Sprintf("Update %s v%s to v%s", common.Build.Name, common.Build.Version, release.Name)

			// Append update type information
			if major {
				title = fmt.Sprintf("%s (Major)", title)
			} else if minor {
				title = fmt.Sprintf("%s (Minor)", title)
			} else if patch {
				title = fmt.Sprintf("%s (Patch)", title)
			}

			// Check update file permissions
			_, err := ui.CheckPermissions(common.Process.Path)
			permitted := err == nil
			if !permitted {
				title = "Missing write permissions for update"
			}
			subitem := version.AddSubMenuItem(title, title)
			enabled := permitted && !major && !minor

			// Update hint icon
			icon := ui.HintIcon(enabled)
			subitem.SetIcon(icon)
			if !enabled {
				subitem.Disable()
			}

			// Update item click
			go func(info common.Info) {
				for {
					<-subitem.ClickedCh

					// Update running binary
					ws := tr.ActiveWorkspace()
					ui.UpdateBinary(ws, info, func() {
						Restart(tr)
					})
				}
			}(*release.Extra)
		}
	}

	// Menu items
	menu = &Menu{}
	systray.AddSeparator()
	for _, entry := range common.Config.TilingIcon {
		action, text := entry[0], entry[1]

		// Separator
		if len(action) == 0 {
			systray.AddSeparator()
			continue
		}

		// Menu item
		var item *systray.MenuItem
		switch action {
		case "toggle":
			item = systray.AddMenuItemCheckbox(text, text, common.Config.TilingEnabled)
			menu.Toggle = item
		case "decoration":
			item = systray.AddMenuItemCheckbox(text, text, common.Config.WindowDecoration)
			menu.Decoration = item
		case "restart":
			item = systray.AddMenuItem(text, text)
		case "exit":
			item = systray.AddMenuItem(text, text)
		default:
			item = systray.AddMenuItem(text, text)
			menu.Actions = append(menu.Actions, item)
		}

		// Menu item action
		go func(action string) {
			for {
				<-item.ClickedCh
				ExecuteAction(action, tr, tr.ActiveWorkspace())
			}
		}(action)
	}
}

func messages(tr *desktop.Tracker) {
	var destination string

	// Request owner of shared session
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Warn("Error initializing tray owner: ", err)
		return
	}
	name := fmt.Sprintf("org.kde.StatusNotifierItem-%d-1", common.Process.Id)
	conn.BusObject().Call("org.freedesktop.DBus.GetNameOwner", 0, name).Store(&destination)
	if len(destination) == 0 {
		log.Warn("Error requesting tray owner: ", name)
		return
	}

	// Monitor method calls in separate session
	conn, err = dbus.ConnectSessionBus()
	if err != nil {
		log.Warn("Error initializing tray methods: ", err)
		return
	}
	call := conn.BusObject().Call("org.freedesktop.DBus.Monitoring.BecomeMonitor", 0, []string{
		fmt.Sprintf("type='method_call',path='/StatusNotifierMenu',interface='com.canonical.dbusmenu',destination='%s'", destination),
		fmt.Sprintf("type='method_call',path='/StatusNotifierItem',interface='org.kde.StatusNotifierItem',destination='%s'", destination),
	}, uint(0))
	if call.Err != nil {
		log.Warn("Error monitoring tray methods: ", call.Err)
		return
	}

	// Listen to channel events
	ch := make(chan *dbus.Message, 10)
	conn.Eavesdrop(ch)

	go func() {
		var iface string
		var method string
		for msg := range ch {
			msg.Headers[2].Store(&iface)
			msg.Headers[3].Store(&method)

			log.Debug(method, " from dbus interface ", iface, " ", msg.Body)

			switch method {
			case "Activate", "SecondaryActivate", "AboutToShow", "AboutToShowGroup":
				clicked = true
				onActivate(tr)
			case "Scroll":
				onPointerScroll(tr, msg.Body[0].(int32), strings.ToLower(msg.Body[1].(string)))
			}
		}
	}()
}

func onExecute(tr *desktop.Tracker, action string) {
	if !common.IsInList(action, []string{"enable", "disable", "toggle", "decoration", "restore", "reset"}) {
		return
	}
	onActivate(tr)
}

func onActivate(tr *desktop.Tracker) {
	ws := tr.ActiveWorkspace()
	al := ws.ActiveLayout()
	mg := al.GetManager()

	if ws.TilingEnabled() {

		// Check toggle item
		if menu.Toggle != nil {
			menu.Toggle.Check()
		}

		// Enable action items
		if menu.Decoration != nil {
			menu.Decoration.Enable()
		}
		for _, item := range menu.Actions {
			item.Enable()
		}
	} else {

		// Uncheck toggle item
		if menu.Toggle != nil {
			menu.Toggle.Uncheck()
		}

		// Disable action items
		if menu.Decoration != nil {
			menu.Decoration.Disable()
		}
		for _, item := range menu.Actions {
			item.Disable()
		}
	}

	if mg.DecorationEnabled() {

		// Check decoration item
		if menu.Decoration != nil {
			menu.Decoration.Check()
		}
	} else {

		// Uncheck decoration item
		if menu.Decoration != nil {
			menu.Decoration.Uncheck()
		}
	}
}

func onPointerClick(tr *desktop.Tracker, pointer store.XPointer) {
	if pointer.Pressed() {
		button = pointer.Button
	}

	// Reset timer
	if click != nil {
		click.Stop()
	}

	// Wait for dbus events
	click = time.AfterFunc(150*time.Millisecond, func() {
		if clicked && button.Left {
			ExecuteAction(common.Config.Systray["click_left"], tr, tr.ActiveWorkspace())
		}
		if clicked && button.Middle {
			ExecuteAction(common.Config.Systray["click_middle"], tr, tr.ActiveWorkspace())
		}
		if clicked && button.Right {
			ExecuteAction(common.Config.Systray["click_right"], tr, tr.ActiveWorkspace())
		}
		clicked = false
	})
}

func onPointerScroll(tr *desktop.Tracker, delta int32, orientation string) {

	// Reset timer
	if click != nil {
		click.Stop()
	}

	// Compress scroll events
	click = time.AfterFunc(150*time.Millisecond, func() {
		switch orientation {
		case "vertical":
			if delta >= 0 {
				ExecuteAction(common.Config.Systray["scroll_down"], tr, tr.ActiveWorkspace())
			} else {
				ExecuteAction(common.Config.Systray["scroll_up"], tr, tr.ActiveWorkspace())
			}
		case "horizontal":
			if delta >= 0 {
				ExecuteAction(common.Config.Systray["scroll_right"], tr, tr.ActiveWorkspace())
			} else {
				ExecuteAction(common.Config.Systray["scroll_left"], tr, tr.ActiveWorkspace())
			}
		}
	})
}

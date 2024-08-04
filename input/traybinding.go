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

	log "github.com/sirupsen/logrus"
)

var (
	clicked bool          // Tray clicked state from dbus
	button  store.XButton // Pointer button state of device
	timer   *time.Timer   // Timer to compress pointer events
	menu    *Menu         // Items collection of systray menu
)

type Menu struct {
	Toggle *systray.MenuItem   // Toggle checkbox item
	Items  []*systray.MenuItem // Menu items for actions
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
	OnExecute(func(action string, desk uint, screen uint) {
		onExecute(tr, action)
	})

	// Attach pointer events
	store.OnPointerUpdate(func(pointer store.XPointer, desk uint, screen uint) {
		onPointerClick(tr, pointer)
	})
}

func items(tr *desktop.Tracker) {
	systray.SetTooltip(fmt.Sprintf("%s - tiling manager", common.Build.Name))
	systray.SetTitle(common.Build.Name)

	// Version checker
	latest := common.VersionToInt(common.Build.Latest)
	current := common.VersionToInt(common.Build.Version)
	title := fmt.Sprintf("%s v%s", common.Build.Name, common.Build.Version)
	if latest > current {
		title = fmt.Sprintf("%s (v%s available)", title, common.Build.Latest)
	}
	version := systray.AddMenuItem(title, title)
	version.SetIcon(common.File.Icon)

	// Menu item hyperlink
	if latest > current {
		go func() {
			for {
				<-version.ClickedCh
				exec.Command("xdg-open", common.Build.Source+"/releases/tag/v"+common.Build.Latest).Start()
			}
		}()
	} else {
		version.Disable()
	}

	// Menu items
	menu = &Menu{}
	systray.AddSeparator()
	for _, m := range common.Config.TilingIcon {
		action, text := m[0], m[1]

		// Separator item
		if len(action) == 0 {
			systray.AddSeparator()
			continue
		}

		// Menu item
		var item *systray.MenuItem
		switch action {
		case "toggle":
			item = systray.AddMenuItemCheckbox(text, text, common.Config.TilingEnabled)
		case "exit":
			item = systray.AddMenuItem(text, text)
		default:
			item = systray.AddMenuItem(text, text)
			menu.Items = append(menu.Items, item)
		}

		// Checkbox item
		if action == "toggle" {
			menu.Toggle = item
		}

		// Menu item action
		go func() {
			for {
				<-item.ClickedCh
				ExecuteAction(action, tr, tr.ActiveWorkspace())
			}
		}()
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
	if !common.IsInList(action, []string{"enable", "disable", "restore", "toggle"}) {
		return
	}
	onActivate(tr)
}

func onActivate(tr *desktop.Tracker) {
	ws := tr.ActiveWorkspace()

	if ws.TilingEnabled() {

		// Check toggle item
		if menu.Toggle != nil {
			menu.Toggle.Check()
		}

		// Enable action items
		for _, item := range menu.Items {
			item.Enable()
		}
	} else {

		// Uncheck toggle item
		if menu.Toggle != nil {
			menu.Toggle.Uncheck()
		}

		// Disable action items
		for _, item := range menu.Items {
			item.Disable()
		}
	}
}

func onPointerClick(tr *desktop.Tracker, pointer store.XPointer) {
	if pointer.Pressed() {
		button = pointer.Button
	}

	// Reset timer
	if timer != nil {
		timer.Stop()
	}

	// Wait for dbus events
	timer = time.AfterFunc(150*time.Millisecond, func() {
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
	if timer != nil {
		timer.Stop()
	}

	// Compress scroll events
	timer = time.AfterFunc(150*time.Millisecond, func() {
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

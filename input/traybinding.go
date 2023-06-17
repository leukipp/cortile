package input

import (
	"fmt"
	"os"
	"time"

	"fyne.io/systray"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/godbus/dbus/v5"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"
	"github.com/leukipp/cortile/store"

	log "github.com/sirupsen/logrus"
)

var (
	clicked bool        // Tray icon clicked flag
	pointer uint16      // Pointer button states of device
	timer   *time.Timer // Timer to compress scroll events
	menu    *Menu       // Items collection of systray menu
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
	OnExecute(func(action string) {
		onExecute(tr, action)
	})

	// Attach pointer events
	store.OnPointerUpdate(func(button uint16) {
		onPointerClick(tr, button)
	})
}

func items(tr *desktop.Tracker) {
	systray.SetTooltip(fmt.Sprintf("%s - tiling manager", common.Build.Name))
	systray.SetTitle(common.Build.Name)

	// Version section
	title := fmt.Sprintf("%s v%s", common.Build.Name, common.Build.Version)
	version := systray.AddMenuItem(title, title)
	version.SetIcon(common.File.Icon)
	version.Disable()

	// Menu section
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

		// Menu item clicked
		go func() {
			for {
				<-item.ClickedCh
				Execute(action, tr)
			}
		}()
	}
}

func messages(tr *desktop.Tracker) {
	var destination string

	// Request owner of shared session
	conn, err := dbus.SessionBus()
	if err != nil {
		log.Warn("Error initializing tray events ", err)
		return
	}
	name := fmt.Sprintf("org.kde.StatusNotifierItem-%d-1", os.Getpid())
	conn.BusObject().Call("org.freedesktop.DBus.GetNameOwner", 0, name).Store(&destination)

	if len(destination) == 0 {
		log.Warn("Error requesting tray owner ", name)
		return
	}

	// Monitor method calls in separate session
	conn, err = dbus.ConnectSessionBus()
	if err != nil {
		log.Warn("Error monitoring tray messages ", err)
		return
	}
	conn.BusObject().Call("org.freedesktop.DBus.Monitoring.BecomeMonitor", 0, []string{
		fmt.Sprintf("type='method_call',path='/StatusNotifierMenu',interface='com.canonical.dbusmenu',destination='%s'", destination),
		fmt.Sprintf("type='method_call',path='/StatusNotifierItem',interface='org.kde.StatusNotifierItem',destination='%s'", destination),
	}, uint(0))

	ch := make(chan *dbus.Message, 10)
	conn.Eavesdrop(ch)

	// Listen to channel events
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
				onPointerScroll(tr, msg.Body[0].(int32), msg.Body[1].(string))
			}
		}
	}()
}

func onExecute(tr *desktop.Tracker, action string) {
	if !common.IsInList(action, []string{"tile", "untile", "toggle"}) {
		return
	}
	onActivate(tr)
}

func onActivate(tr *desktop.Tracker) {
	ws := tr.ActiveWorkspace()

	if !ws.IsEnabled() {

		// Uncheck toggle item
		if menu.Toggle != nil {
			menu.Toggle.Uncheck()
		}

		// Disable action items
		for _, item := range menu.Items {
			item.Disable()
		}
	} else {

		// Check toggle item
		if menu.Toggle != nil {
			menu.Toggle.Check()
		}

		// Enable action items
		for _, item := range menu.Items {
			item.Enable()
		}
	}
}

func onPointerClick(tr *desktop.Tracker, button uint16) {
	if button != 0 {
		pointer = button
	}

	// Wait for dbus event
	time.AfterFunc(50*time.Millisecond, func() {
		if clicked {
			switch pointer {
			case pointer & xproto.ButtonMask1:
				Execute(common.Config.Systray["click_left"], tr)
			case pointer & xproto.ButtonMask2:
				Execute(common.Config.Systray["click_middle"], tr)
			case pointer & xproto.ButtonMask3:
				Execute(common.Config.Systray["click_right"], tr)
			}
			clicked = false
		}
	})
}

func onPointerScroll(tr *desktop.Tracker, delta int32, orientation string) {
	if timer != nil {
		timer.Stop()
	}

	// Compress scroll events
	timer = time.AfterFunc(50*time.Millisecond, func() {
		switch orientation {
		case "vertical":
			if delta >= 0 {
				Execute(common.Config.Systray["scroll_down"], tr)
			} else {
				Execute(common.Config.Systray["scroll_up"], tr)
			}
		case "horizontal":
			if delta >= 0 {
				Execute(common.Config.Systray["scroll_right"], tr)
			} else {
				Execute(common.Config.Systray["scroll_left"], tr)
			}
		}
	})
}

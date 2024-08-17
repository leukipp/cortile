package input

import (
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"golang.org/x/exp/maps"

	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgbutil/ewmh"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"

	"github.com/leukipp/cortile/v2/common"
	"github.com/leukipp/cortile/v2/desktop"
	"github.com/leukipp/cortile/v2/store"

	log "github.com/sirupsen/logrus"
)

var (
	iface   string           // Dbus interface name
	opath   dbus.ObjectPath  // Dbus object path
	props   *prop.Properties // Dbus properties
	methods *Methods         // Dbus methods
)

type Methods struct {
	Naming  map[string][]string // Method and arguments names
	Tracker *desktop.Tracker    // Workspace tracker instance
}

func (m Methods) ActionExecute(name string, desktop int32, screen int32) (string, *dbus.Error) {
	success := false

	// Execute action
	ws := m.Tracker.WorkspaceAt(uint(desktop), uint(screen))
	if ws != nil {
		success = ExecuteAction(name, m.Tracker, ws)
	}

	// Return result
	result := common.Map{"Success": success}

	return dataMap("Result", "ActionExecute", result), nil
}

func (m Methods) WindowActivate(id int32) (string, *dbus.Error) {
	success := false

	// Activate window
	if c, ok := m.Tracker.Clients[xproto.Window(id)]; ok {
		store.ActiveWindowSet(store.X, c.Window)
		success = true
	}

	// Return result
	result := common.Map{"Success": success}

	return dataMap("Result", "WindowActivate", result), nil
}

func (m Methods) WindowToPosition(id int32, x int32, y int32) (string, *dbus.Error) {
	success := false

	// Move window to position
	valid := x >= 0 && y >= 0
	if c, ok := m.Tracker.Clients[xproto.Window(id)]; ok && valid {
		ewmh.MoveWindow(store.X, c.Window.Id, int(x), int(y))
		store.Pointer.Press()
		success = true
	}

	// Return result
	result := common.Map{"Success": success}

	return dataMap("Result", "WindowToPosition", result), nil
}

func (m Methods) WindowToDesktop(id int32, desktop int32) (string, *dbus.Error) {
	success := false

	// Move window to desktop
	valid := desktop >= 0 && uint(desktop) < store.Workplace.DesktopCount
	if c, ok := m.Tracker.Clients[xproto.Window(id)]; ok && valid {
		success = c.MoveToDesktop(uint32(desktop))
	}

	// Return result
	result := common.Map{"Success": success}

	return dataMap("Result", "WindowToDesktop", result), nil
}

func (m Methods) WindowToScreen(id int32, screen int32) (string, *dbus.Error) {
	success := false

	// Move window to screen
	valid := screen >= 0 && uint(screen) < store.Workplace.ScreenCount
	if c, ok := m.Tracker.Clients[xproto.Window(id)]; ok && valid {
		success = c.MoveToScreen(uint32(screen))
	}

	// Return result
	result := common.Map{"Success": success}

	return dataMap("Result", "WindowToScreen", result), nil
}

func (m Methods) DesktopSwitch(desktop int32) (string, *dbus.Error) {
	success := false

	// Switch current desktop
	valid := desktop >= 0 && uint(desktop) < store.Workplace.DesktopCount
	if valid {
		store.CurrentDesktopSet(store.X, uint(desktop))
		success = true
	}

	// Return result
	result := common.Map{"Success": success}

	return dataMap("Result", "DesktopSwitch", result), nil
}

func (m Methods) Introspection() []introspect.Method {
	typ := reflect.TypeOf(m)
	ims := make([]introspect.Method, 0, typ.NumMethod())

	for i := 0; i < typ.NumMethod(); i++ {
		if typ.Method(i).PkgPath != "" {
			continue
		}

		// Validate return types
		mt := typ.Method(i).Type
		if mt.NumOut() == 0 || mt.Out(mt.NumOut()-1) != reflect.TypeOf(&dbus.Error{}) {
			continue
		}

		// Introspect method
		im := introspect.Method{
			Name:        typ.Method(i).Name,
			Annotations: make([]introspect.Annotation, 0),
			Args:        make([]introspect.Arg, 0, mt.NumIn()+mt.NumOut()-2),
		}

		// Arguments in
		for j := 1; j < mt.NumIn(); j++ {
			styp := dbus.SignatureOfType(mt.In(j)).String()
			im.Args = append(im.Args, introspect.Arg{Name: m.Naming[im.Name][j-1], Type: styp, Direction: "in"})
		}

		// Arguments out
		for j := 0; j < mt.NumOut()-1; j++ {
			styp := dbus.SignatureOfType(mt.Out(j)).String()
			im.Args = append(im.Args, introspect.Arg{Name: "json", Type: styp, Direction: "out"})
		}

		ims = append(ims, im)
	}

	return ims
}

func BindDbus(tr *desktop.Tracker) {

	// Export interfaces
	if !common.HasFlag("disable-dbus-interface") {
		go export(tr)
	}

	// Bind event channel
	go event(tr.Channels.Event, tr)

	// Attach execute events
	OnExecute(func(action string, desktop uint, screen uint) {
		SetProperty("Action", struct {
			Name     string
			Location store.Location
		}{
			Name:     action,
			Location: store.Location{Desktop: desktop, Screen: screen},
		})
	})

	// Attach pointer events
	store.OnPointerUpdate(func(pointer store.XPointer, desktop uint, screen uint) {
		SetProperty("Pointer", struct {
			Device   store.XPointer
			Location store.Location
		}{
			Device:   pointer,
			Location: store.Location{Desktop: desktop, Screen: screen},
		})
	})
}

func event(ch chan string, tr *desktop.Tracker) {
	for {
		switch <-ch {
		case "clients_change":
			SetProperty("Clients", common.Map{"Values": maps.Values(tr.Clients)})
		case "workspaces_change":
			SetProperty("Workspaces", common.Map{"Values": maps.Values(tr.Workspaces)})
		case "workplace_change":
			SetProperty("Workplace", *store.Workplace)
		case "windows_change":
			SetProperty("Windows", *store.Windows)
		case "corner_change":
			for _, hc := range store.Workplace.Displays.Corners {
				if !hc.Active {
					continue
				}
				SetProperty("Corner", struct {
					Name     string
					Location store.Location
				}{
					Name:     hc.Name,
					Location: tr.ActiveWorkspace().Location,
				})
			}
		}
	}
}

func connect() (*dbus.Conn, error) {
	hostname := strings.Join(common.ReverseList(strings.Split(common.Source.Hostname, ".")), ".")
	repository := strings.Replace(common.Source.Repository, "/", ".", -1)

	// Init interface and path
	iface = fmt.Sprintf("%s.%s", hostname, repository)
	opath = dbus.ObjectPath(fmt.Sprintf("/%s", strings.Replace(iface, ".", "/", -1)))

	// Init session bus
	return dbus.ConnectSessionBus()
}

func export(tr *desktop.Tracker) {
	conn, err := connect()
	if err != nil {
		log.Warn("Error initializing dbus server: ", err)
		return
	}
	defer conn.Close()

	// Request dbus name
	reply, err := conn.RequestName(iface, dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		log.Warn("Error requesting dbus name ", iface, ": ", err)
		return
	}

	// Export dbus properties
	mapping := map[string]interface{}{
		"Process":       structToMap(common.Process),
		"Build":         structToMap(common.Build),
		"Source":        structToMap(common.Source),
		"Arguments":     structToMap(common.Args),
		"Configuration": structToMap(common.Config),
		"Workspaces":    common.Map{},
		"Workplace":     common.Map{},
		"Windows":       common.Map{},
		"Clients":       common.Map{},
		"Pointer":       common.Map{},
		"Action":        common.Map{},
		"Corner":        common.Map{},
		"Disconnect":    common.Map{},
	}
	properties := map[string]*prop.Prop{}
	for name, value := range mapping {
		properties[name] = &prop.Prop{
			Value:    value,
			Emit:     prop.EmitTrue,
			Writable: len(value.(common.Map)) == 0,
		}
	}
	props, err = prop.Export(conn, opath, prop.Map{iface: properties})
	if err != nil {
		log.Warn("Error exporting dbus properties: ", err)
		return
	}

	// Export dbus methods
	methods = &Methods{
		Naming: map[string][]string{
			"ActionExecute":    {"name", "desktop", "screen"},
			"WindowActivate":   {"id"},
			"WindowToPosition": {"id", "x", "y"},
			"WindowToDesktop":  {"id", "desktop"},
			"WindowToScreen":   {"id", "screen"},
			"DesktopSwitch":    {"desktop"},
		},
		Tracker: tr,
	}
	err = conn.Export(methods, opath, iface)
	if err != nil {
		log.Warn("Error exporting dbus methods: ", err)
		return
	}

	// Export dbus interfaces
	intro := introspect.NewIntrospectable(&introspect.Node{
		Name: string(opath),
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			{
				Name:       iface,
				Methods:    methods.Introspection(),
				Properties: props.Introspection(iface),
			},
		},
	})
	err = conn.Export(intro, opath, "org.freedesktop.DBus.Introspectable")
	if err != nil {
		log.Warn("Error exporting dbus interfaces: ", err)
		return
	}

	select {}
}

func Introspect() map[string][]string {
	conn, err := connect()
	if err != nil {
		return map[string][]string{}
	}
	defer conn.Close()

	// Call introspect method
	node, err := introspect.Call(conn.Object(iface, opath))
	if err != nil {
		return map[string][]string{}
	}

	// Iterate node interfaces
	methods := []string{}
	properties := []string{}
	for _, item := range node.Interfaces {
		if item.Name != iface {
			continue
		}

		// Get dbus methods
		for _, method := range item.Methods {
			description := method.Name
			for _, arg := range method.Args {
				if arg.Direction != "in" {
					continue
				}
				switch arg.Type {
				case "s":
					description += fmt.Sprintf(" str:%s", arg.Name)
				case "i":
					description += fmt.Sprintf(" int:%s", arg.Name)
				}
			}
			methods = append(methods, description)
		}

		// Get dbus properties
		for _, property := range item.Properties {
			properties = append(properties, property.Name)
		}
	}
	sort.Strings(methods)
	sort.Strings(properties)

	return map[string][]string{
		"Methods":    methods,
		"Properties": properties,
	}
}

func Method(name string, args []string) {
	conn, err := connect()
	if err != nil {
		fatal("Error initializing dbus server", err)
	}
	defer conn.Close()

	// Convert arguments
	variants := make([]interface{}, len(args))
	for i, value := range args {
		integer, err := strconv.Atoi(value)
		if err == nil {
			variants[i] = dbus.MakeVariant(integer)
		} else {
			variants[i] = dbus.MakeVariant(value)
		}
	}

	// Call dbus method
	call := conn.Object(iface, opath).Call(fmt.Sprintf("%s.%s", iface, name), 0, variants...)
	if call.Err != nil {
		fatal("Error calling dbus method", call.Err)
	}

	// Print reply
	var reply string
	call.Store(&reply)
	fmt.Println(reply)
}

func Property(name string) {
	conn, err := connect()
	if err != nil {
		fatal("Error initializing dbus server", err)
	}
	defer conn.Close()

	// Convert arguments
	variants := []interface{}{dbus.MakeVariant(iface), dbus.MakeVariant(name)}

	// Receive dbus property
	call := conn.Object(iface, opath).Call("org.freedesktop.DBus.Properties.Get", 0, variants...)
	if call.Err != nil {
		fatal("Error receiving dbus property", call.Err)
	}

	// Print reply
	var reply dbus.Variant
	call.Store(&reply)
	print("Property", name, variantToMap(reply))
}

func Listen(args []string) {
	conn, err := connect()
	if err != nil {
		fatal("Error initializing dbus server", err)
	}
	defer conn.Close()

	// Monitor property changes and method calls
	call := conn.BusObject().Call("org.freedesktop.DBus.Monitoring.BecomeMonitor", 0, []string{
		fmt.Sprintf("type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged',path='%s'", opath),
		fmt.Sprintf("type='method_call',interface='%s',path='%s'", iface, opath),
	}, uint(0))
	if call.Err != nil {
		fatal("Error becoming dbus monitor", call.Err)
	}

	// Listen to channel events
	ch := make(chan *dbus.Message, 10)
	conn.Eavesdrop(ch)

	var method string
	for msg := range ch {
		msg.Headers[3].Store(&method)

		// Print reply
		switch method {
		case "PropertiesChanged":
			typ := "Property"
			for name, variant := range msg.Body[1].(map[string]dbus.Variant) {
				filter := fmt.Sprintf("%s:%s", typ, name)
				if len(args) == 0 || common.IsInList(filter, args) {
					print(typ, name, variantToMap(variant))
				}
			}
		default:
			typ := "Method"
			filter := fmt.Sprintf("%s:%s", typ, method)
			if len(args) == 0 || common.IsInList(filter, args) {
				print(typ, method, common.Map{"Body": bodyToString(msg.Body)})
			}
		}
	}
}

func Disconnect() {
	SetProperty("Disconnect", struct {
		Event string
	}{
		Event: "exit",
	})
}

func GetProperty(name string) common.Map {
	if props == nil {
		return common.Map{}
	}
	variant := dbus.MakeVariant(props.GetMust(iface, name))
	return variantToMap(variant)
}

func SetProperty(name string, obj interface{}) {
	if props == nil {
		return
	}
	variant := dbus.MakeVariant(structToMap(obj))
	props.SetMust(iface, name, variant)
}

func variantToMap(variant dbus.Variant) (value common.Map) {
	variant.Store(&value)
	return value
}

func structToMap(obj interface{}) (value common.Map) {
	data, err := json.Marshal(obj)
	if err != nil {
		return value
	}
	json.Unmarshal(data, &value)
	return value
}

func mapToString(obj interface{}) string {
	data, err := json.Marshal(obj)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func bodyToString(obj interface{}) string {
	r := strings.NewReplacer(":", "", "[", "", "]", "", "\"", "")
	return r.Replace(fmt.Sprint(obj))
}

func dataMap(typ string, name string, data common.Map) string {
	time := time.Now().UnixMilli()
	process := common.Process.Id
	return mapToString(common.Map{"Process": process, "Time": time, "Type": typ, "Name": name, "Data": data})
}

func print(typ string, name string, data common.Map) {
	fmt.Println(dataMap(typ, name, data))
}

func fatal(msg string, err error) {
	print("Error", "Fatal", common.Map{"Message": fmt.Sprintf("%s: %s", msg, err)})

	// exit with success error code
	os.Exit(0)
}

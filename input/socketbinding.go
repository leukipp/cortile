package input

import (
	"encoding/json"
	"net"
	"os"
	"strings"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"

	log "github.com/sirupsen/logrus"
)

type Message[T any] struct {
	Type string // Socket message type
	Name string // Socket message name
	Data T      // Socket message data
}

func BindSocket(tr *desktop.Tracker) {

	// Create a unix domain socket listener
	listener, err := net.Listen("unix", common.Args.Sock+".in")
	if err != nil {
		os.Remove(common.Args.Sock + ".in")
		log.Warn("Listener connection error: ", err)
		return
	}
	go listen(listener, tr)
}

func NotifySocket[T any](m Message[T]) {
	if _, err := os.Stat(common.Args.Sock + ".out"); os.IsNotExist(err) {
		return
	}

	// Create a unix domain socket dialer
	dialer, err := net.Dial("unix", common.Args.Sock+".out")
	if err != nil {
		os.Remove(common.Args.Sock + ".out")
		log.Warn("Dealer connection error: ", err)
		return
	}

	// Parse outgoing data
	data, err := json.Marshal(m)
	if err != nil {
		log.Warn("Dealer parse error: ", err)
		return
	}

	msg := string(data)
	log.Info("Send socket message ", common.Truncate(msg, 100), "...")

	// Write outgoing data
	_, err = dialer.Write([]byte(msg))
	if err != nil {
		log.Warn("Dealer write error: ", err)
	}
}

func listen(listener net.Listener, tr *desktop.Tracker) {
	for {

		// Listen for incoming data
		connection, err := listener.Accept()
		if err != nil {
			log.Warn("Listener accept error: ", err)
			return
		}

		// Read incoming data
		data := make([]byte, 4096)
		n, err := connection.Read(data)
		if err != nil {
			log.Warn("Listener read error: ", err)
			return
		}

		msg := strings.TrimSpace(string(data[:n]))
		log.Info("Receive socket message ", common.Truncate(msg, 100), "...")

		// Parse incoming data
		var kv map[string]string
		err = json.Unmarshal(data[:n], &kv)
		if err != nil {
			log.Warn("Listener parse error: ", err)
			return
		}

		// Execute action
		if v, ok := kv["Action"]; ok {
			Execute(v, "current", tr)
		}

		// Query state
		if v, ok := kv["State"]; ok {
			Query(v, tr)
		}
	}
}

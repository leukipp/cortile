package input

import (
	"net"
	"os"
	"strings"

	"github.com/leukipp/cortile/common"
	"github.com/leukipp/cortile/desktop"

	log "github.com/sirupsen/logrus"
)

func BindSocket(tr *desktop.Tracker) {
	os.Remove(common.Args.Sock)

	// Create a unix domain socket listener
	listener, err := net.Listen("unix", common.Args.Sock)
	if err != nil {
		log.Error("Listener connection error: ", err)
		return
	}
	go listen(listener, tr)
}

func listen(listener net.Listener, tr *desktop.Tracker) {
	for {

		// Listen for incoming data
		connection, err := listener.Accept()
		if err != nil {
			log.Error("Listener accept error: ", err)
			return
		}

		// Read incoming data
		data := make([]byte, 4096)
		n, err := connection.Read(data)
		if err != nil {
			log.Error("Listener read error: ", err)
		}

		msg := string(data[:n])
		log.Info("Receive socket message \"", msg, "\"")

		// Check socket message
		if strings.HasPrefix(msg, "command:") {
			cmd := strings.TrimSpace(strings.Split(msg, "command:")[1])

			// Execute command
			Execute(cmd, tr)
		}
	}
}

func Notify(msg string) {

	// Create a unix domain socket dialer
	dialer, err := net.Dial("unix", common.Args.Sock)
	if err != nil {
		log.Error("Dealer connection error: ", err)
		return
	}

	log.Info("Send socket message \"", msg, "\"")

	// Write outgoing data
	_, err = dialer.Write([]byte(msg))
	if err != nil {
		log.Error("Dealer write error: ", err)
	}
}

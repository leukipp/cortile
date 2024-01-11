package input

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/leukipp/cortile/v2/desktop"
)

func BindSignal(tr *desktop.Tracker) {
	ch := make(chan os.Signal, 1)

	// Bind signal channel
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go exit(ch, tr)

	// Bind action channel
	go action(tr.Action, tr)
}

func exit(ch chan os.Signal, tr *desktop.Tracker) {
	<-ch
	Execute("exit", "current", tr)
}

func action(ch chan string, tr *desktop.Tracker) {
	for {
		Execute(<-ch, "current", tr)
	}
}

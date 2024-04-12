package input

import (
	"os"
	"syscall"

	"os/signal"

	"github.com/leukipp/cortile/v2/desktop"
)

func BindSignal(tr *desktop.Tracker) {
	ch := make(chan os.Signal, 1)

	// Bind signal channel
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go exit(ch, tr)
}

func exit(ch chan os.Signal, tr *desktop.Tracker) {
	<-ch
	ExecuteAction("exit", tr, tr.ActiveWorkspace())
}

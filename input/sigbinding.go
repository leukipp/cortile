package input

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/leukipp/cortile/desktop"
)

func BindSig(tr *desktop.Tracker) {
	c := make(chan os.Signal, 1)

	// Bind signal events
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go exit(c, tr)
}

func exit(c chan os.Signal, tr *desktop.Tracker) {
	<-c
	Execute("exit", tr)
}

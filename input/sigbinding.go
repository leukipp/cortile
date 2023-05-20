package input

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/leukipp/cortile/desktop"
)

func BindSig(t *desktop.Tracker) {
	c := make(chan os.Signal, 1)

	// Bind signal events
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go exit(c, t)
}

func exit(c chan os.Signal, t *desktop.Tracker) {
	<-c
	Execute("untile", t)
	os.Exit(1)
}

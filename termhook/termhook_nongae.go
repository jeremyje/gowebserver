// +build !appengine

package termhook

import (
	"os"
	"os/signal"
	"syscall"
)

func addTerminatingSignals(c chan<- os.Signal) {
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
}

// +build appengine

package termhook

import (
	"os"
	"os/signal"
)

func addTerminatingSignals(c chan<- os.Signal) {
	signal.Notify(c, os.Interrupt)
}

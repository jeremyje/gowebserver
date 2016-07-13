package termhook

import (
	"os"
)

type SignalCallback func(os.Signal)

// AddWithSignal enqueues a callback that will be run when the app is terminated.
func AddWithSignal(callback SignalCallback) {
	globalSignalManager.addCallback(callback)
}

// Add enqueues a simple func() callback to run when the app is terminated.
func Add(callback func()) {
	AddWithSignal(func(sig os.Signal) {
		callback()
	})
}

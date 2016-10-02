package termhook

import (
	"os"
)

type signalManager struct {
	intest       *atomicBool
	isclosed     *atomicBool
	channel      chan os.Signal
	callbackList []SignalCallback
	testchan     chan bool
}

var globalSignalManager *signalManager

func newSignalManager() *signalManager {
	manager := &signalManager{
		intest:       newAtomicBool(),
		channel:      make(chan os.Signal, 1),
		isclosed:     newAtomicBool(),
		callbackList: []SignalCallback{},
		testchan:     make(chan bool),
	}

	addTerminatingSignals(manager.channel)
	return manager
}

func (this *signalManager) startListening() {
	go func() {
		if !this.isclosed.get() {
			for sig := range this.channel {
				for _, callback := range this.callbackList {
					callback(sig)
				}
				if !this.intest.get() {
					os.Exit(0xf)
				} else {
					this.testchan <- true
				}
			}
		}
	}()
}

func (this *signalManager) stopListening() {
	this.isclosed.set(true)
	close(this.channel)
}

func (this *signalManager) addCallback(callback SignalCallback) {
	this.callbackList = append(this.callbackList, callback)
}

func init() {
	globalSignalManager = newSignalManager()
	globalSignalManager.startListening()
}

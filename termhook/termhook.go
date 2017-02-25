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

func (sm *signalManager) startListening() {
	go func() {
		if !sm.isclosed.get() {
			for sig := range sm.channel {
				for _, callback := range sm.callbackList {
					callback(sig)
				}
				if !sm.intest.get() {
					os.Exit(0xf)
				} else {
					sm.testchan <- true
				}
			}
		}
	}()
}

func (sm *signalManager) stopListening() {
	sm.isclosed.set(true)
	close(sm.channel)
}

func (sm *signalManager) addCallback(callback SignalCallback) {
	sm.callbackList = append(sm.callbackList, callback)
}

func init() {
	globalSignalManager = newSignalManager()
	globalSignalManager.startListening()
}

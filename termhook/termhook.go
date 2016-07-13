package termhook

import (
	"os"
)

type signalManager struct {
	channel      chan os.Signal
	callbackList []SignalCallback
}

var globalSignalManager *signalManager

func newSignalManager() *signalManager {
	manager := &signalManager{
		channel:      make(chan os.Signal, 1),
		callbackList: []SignalCallback{},
	}

	addTerminatingSignals(manager.channel)
	return manager
}

func (this *signalManager) startListening() {
	go func() {
		if this.channel != nil {
			for sig := range this.channel {
				for _, callback := range this.callbackList {
					callback(sig)
				}
			}
		}
	}()
}

func (this *signalManager) stopListening() {
	close(this.channel)
	this.channel = nil
}

func (this *signalManager) addCallback(callback SignalCallback) {
	this.callbackList = append(this.callbackList, callback)
}

func init() {
	globalSignalManager = newSignalManager()
	globalSignalManager.startListening()
}

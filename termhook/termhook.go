package termhook

import (
	"log"
	"os"
)

type signalManager struct {
	intest       bool
	channel      chan os.Signal
	callbackList []SignalCallback
}

var globalSignalManager *signalManager

func newSignalManager() *signalManager {
	manager := &signalManager{
		intest:       false,
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
				log.Printf("Sigterming: %t", this.intest)
				if !this.intest {
					os.Exit(0xf)
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

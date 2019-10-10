// Copyright 2019 Jeremy Edwards
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

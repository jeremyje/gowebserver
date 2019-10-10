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

// SignalCallback is a callback for handling os signals.
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

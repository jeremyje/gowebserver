package termhook

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"log"
)

func TestAddCallback(t *testing.T) {
	assert := assert.New(t)

	manager := newSignalManagerForTest()

	signalCaught := false

	manager.addCallback(func(sig os.Signal) {
		signalCaught = true
	})
	manager.startListening()

	assert.False(signalCaught, "signalCaught should be false.")

	// Send 2 signals to ensure the first one gets through, this is flaky, needs better fix.
	manager.channel <- os.Interrupt
	manager.channel <- os.Interrupt
	assert.True(signalCaught, "signalCaught should be true.")
}

func TestMultipleCallbacks(t *testing.T) {
	assert := assert.New(t)

	manager := newSignalManagerForTest()

	signalCaughtOne := false
	signalCaughtTwo := false

	manager.addCallback(func(sig os.Signal) {
		signalCaughtOne = true
	})
	manager.addCallback(func(sig os.Signal) {
		signalCaughtTwo = true
	})

	manager.startListening()

	assert.False(signalCaughtOne, "signalCaughtOne should be false.")
	assert.False(signalCaughtTwo, "signalCaughtTwo should be false.")

	// Send 2 signals to ensure the first one gets through, this is flaky, needs better fix.
	manager.channel <- os.Interrupt
	manager.channel <- os.Interrupt
	assert.True(signalCaughtOne, "signalCaughtOne should be true.")
	assert.True(signalCaughtTwo, "signalCaughtTwo should be true.")
}

func TestStopListening(t *testing.T) {
	assert := assert.New(t)

	manager := newSignalManagerForTest()

	signalCaught := false

	manager.addCallback(func(sig os.Signal) {
		signalCaught = true
	})
	manager.startListening()
	manager.stopListening()

	assert.Nil(manager.channel, "manager.channel should be nil.")
}

func simulateSignal() {
	// Tell the signal manager to not kill the app.
	globalSignalManager.intest = true
	log.Printf("Set Val (race condition?): %t", globalSignalManager.intest)
	// Send 2 signals to ensure the first one gets through, this is flaky, needs better fix.
	globalSignalManager.channel <- os.Interrupt
	globalSignalManager.channel <- os.Interrupt
	globalSignalManager.channel <- os.Interrupt
	globalSignalManager.channel <- os.Interrupt
	globalSignalManager.channel <- os.Interrupt
	globalSignalManager.channel <- os.Interrupt
	globalSignalManager.channel <- os.Interrupt
	globalSignalManager.channel <- os.Interrupt
	globalSignalManager.channel <- os.Interrupt
	globalSignalManager.channel <- os.Interrupt
	globalSignalManager.channel <- os.Interrupt
	globalSignalManager.channel <- os.Interrupt
}

func TestGlobalSignalCallbacks(t *testing.T) {
	assert := assert.New(t)

	signalCaught := false

	AddWithSignal(func(sig os.Signal) {
		signalCaught = true
	})

	assert.False(signalCaught)

	simulateSignal()

	assert.True(signalCaught)
}

func ExampleAdd() {
	signalCaught := false

	Add(func() {
		signalCaught = true
	})
	simulateSignal()

	fmt.Printf("Signal Caught: %t", signalCaught)
	// Output: Signal Caught: true
}

func ExampleAddWithSignal() {
	signalCaught := false

	AddWithSignal(func(sig os.Signal) {
		signalCaught = true
	})
	simulateSignal()

	fmt.Printf("Signal Caught: %t", signalCaught)
	// Output: Signal Caught: true
}

func newSignalManagerForTest() *signalManager {
	m := newSignalManager()
	m.intest = true
	return m
}

package termhook

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestAddCallback(t *testing.T) {
	assert := assert.New(t)

	manager := newSignalManagerForTest()

	signalCaught := newAtomicBool()

	manager.addCallback(func(sig os.Signal) {
		signalCaught.set(true)
	})
	manager.startListening()

	assert.False(signalCaught.get(), "signalCaught should be false.")

	// Send 2 signals to ensure the first one gets through, this is flaky, needs better fix.
	manager.channel <- os.Interrupt
	manager.channel <- os.Interrupt
	assert.True(signalCaught.get(), "signalCaught should be true.")
}

func TestMultipleCallbacks(t *testing.T) {
	assert := assert.New(t)

	manager := newSignalManagerForTest()

	signalCaughtOne := newAtomicBool()
	signalCaughtTwo := newAtomicBool()

	manager.addCallback(func(sig os.Signal) {
		signalCaughtOne.set(true)
	})
	manager.addCallback(func(sig os.Signal) {
		signalCaughtTwo.set(true)
	})

	manager.startListening()

	assert.False(signalCaughtOne.get(), "signalCaughtOne should be false.")
	assert.False(signalCaughtTwo.get(), "signalCaughtTwo should be false.")

	// Send 2 signals to ensure the first one gets through, this is flaky, needs better fix.
	manager.channel <- os.Interrupt
	manager.channel <- os.Interrupt
	assert.True(signalCaughtOne.get(), "signalCaughtOne should be true.")
	assert.True(signalCaughtTwo.get(), "signalCaughtTwo should be true.")
}

func TestStopListening(t *testing.T) {
	assert := assert.New(t)

	manager := newSignalManagerForTest()

	signalCaught := newAtomicBool()

	manager.addCallback(func(sig os.Signal) {
		signalCaught.set(true)
	})
	manager.startListening()
	manager.stopListening()

	assert.True(manager.isclosed.get(), "manager.channel should be nil.")
}

func simulateSignal() {
	// Tell the signal manager to not kill the app.
	globalSignalManager.intest.set(true)
	log.Printf("Set Val (race condition?): %t", globalSignalManager.intest.get())
	globalSignalManager.channel <- os.Interrupt
	<-globalSignalManager.testchan
}

func TestGlobalSignalCallbacks(t *testing.T) {
	assert := assert.New(t)

	signalCaught := newAtomicBool()

	AddWithSignal(func(sig os.Signal) {
		signalCaught.set(true)
	})

	assert.False(signalCaught.get())

	simulateSignal()

	assert.True(signalCaught.get())
}

func ExampleAdd() {
	signalCaught := newAtomicBool()

	Add(func() {
		signalCaught.set(true)
	})
	simulateSignal()

	fmt.Printf("Signal Caught: %t", signalCaught.get())
	// Output: Signal Caught: true
}

func ExampleAddWithSignal() {
	signalCaught := newAtomicBool()

	AddWithSignal(func(sig os.Signal) {
		signalCaught.set(true)
	})
	simulateSignal()

	fmt.Printf("Signal Caught: %t", signalCaught.get())
	// Output: Signal Caught: true
}

func newSignalManagerForTest() *signalManager {
	m := newSignalManager()
	m.intest.set(true)
	return m
}

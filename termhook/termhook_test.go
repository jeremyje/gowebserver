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

	simulateSignalOnManager(manager)

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

	simulateSignalOnManager(manager)

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

	assert.True(manager.isclosed.get(), "manager.isclosed should be true")
	assert.False(signalCaught.get(), "signal caught should not be set.")
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

func simulateSignal() {
	simulateSignalOnManager(globalSignalManager)
}

func simulateSignalOnManager(manager *signalManager) {
	// Tell the signal manager to not kill the app.
	manager.intest.set(true)
	log.Printf("Set Val (race condition?): %t", manager.intest.get())
	manager.channel <- os.Interrupt
	<-manager.testchan
}

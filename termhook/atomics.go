package termhook

import "sync/atomic"

type atomicBool struct{ flag int32 }

func (b *atomicBool) set(value bool) {
	var i int32 = 0
	if value {
		i = 1
	}
	atomic.StoreInt32(&(b.flag), int32(i))
}

func (b *atomicBool) get() bool {
	if atomic.LoadInt32(&(b.flag)) != 0 {
		return true
	}
	return false
}

func newAtomicBool() *atomicBool {
	return new(atomicBool)
}

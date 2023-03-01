package lockModel

import (
	"kiteu/util"
	"sync"
)

type RWLock struct {
	lock sync.RWMutex
}

func (rw *RWLock) Read(r util.Callback) {
	rw.lock.RLock()
	util.SafeCallback(r)
	rw.lock.RUnlock()
}
func (rw *RWLock) Write(w util.Callback) {
	rw.lock.Lock()
	util.SafeCallback(w)
	rw.lock.Unlock()
}

type Lock struct {
	lock sync.Mutex
}

func (l *Lock) Do(f util.Callback) {
	l.lock.Lock()
	util.SafeCallback(f)
	l.lock.Unlock()
}

package collapsar

import (
	"encoding/gob"
	"io"
	"sync"
	"sync/atomic"

	"github.com/pkg/errors"
)

type writer struct {
	state     uint32
	max, size int32
	onFull    chan struct{}
	mu        sync.Mutex
	entries   [][]byte
}

func (a *writer) onFullTrigger() {
	if atomic.LoadUint32(&a.state) == 0 || atomic.LoadInt32(&a.max) != atomic.LoadInt32(&a.size) {
		return
	}
	_ = a.Close()
}

func (a *writer) Encode(w io.Writer) error {
	if atomic.LoadUint32(&a.state) == 0 {
		return errors.New("collapsar writer has closed")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	return gob.NewEncoder(w).Encode(Binary{
		Size:    atomic.LoadInt32(&a.size),
		Entries: a.entries,
	})
}

func (a *writer) Add(entry []byte) error {
	if atomic.LoadUint32(&a.state) == 0 {
		return errors.New("collapsar writer has closed")
	}
	if atomic.LoadInt32(&a.max) < atomic.LoadInt32(&a.size) {
		return errors.New("reach the max")
	}
	// put to entries
	a.entries[atomic.AddInt32(&a.size, 1)-1] = entry
	a.onFullTrigger()
	return nil
}

func (a *writer) Wait() {
	if atomic.LoadUint32(&a.state) == 0 {
		return
	}
	<-a.onFull
}

func (a *writer) Close() error {
	if atomic.LoadUint32(&a.state) == 0 {
		return errors.New("repeat close")
	}
	atomic.StoreUint32(&a.state, 0)

	a.mu.Lock()
	close(a.onFull)
	a.mu.Unlock()

	return nil
}

func NewWriter(max int32) Writer {
	return &writer{
		state:   1,
		max:     max,
		size:    0,
		onFull:  make(chan struct{}),
		entries: make([][]byte, max),
	}
}

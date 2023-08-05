package client

import (
	"sync/atomic"

	"github.com/pkg/errors"
)

const DefaultWatchBufSize uint32 = 8

var (
	ErrWatcherClosed = errors.New("watcher has closed")
)

type WatchValue struct {
	seq       uint64
	Timestamp int64
	TTL       uint32
	Key       []byte
	Value     []byte
}

type Watcher struct {
	close        atomic.Bool
	ignoreErrors bool
	prefixWatch  bool

	bufSize uint32
	// if prefixWatch equal true, store prefix
	key       []byte
	namespace *string
	ch        chan *WatchValue
}

func (w *Watcher) Close() error {
	if w.close.Load() {
		return ErrWatcherClosed
	}
	w.close.Store(true)
	close(w.ch)
	return nil
}

func (w *Watcher) Notify() (chan *WatchValue, error) {
	if w.close.Load() {
		return nil, ErrWatcherClosed
	}
	return w.ch, nil
}

type WatcherOption func(*Watcher)

func NewWatcher(key []byte, opts ...WatcherOption) *Watcher {
	w := &Watcher{
		key: key,
	}

	for _, opt := range opts {
		opt(w)
	}

	if w.bufSize == 0 {
		w.bufSize = DefaultWatchBufSize
	}

	w.ch = make(chan *WatchValue, w.bufSize)

	return w
}

func WatchWithPrefix() WatcherOption {
	return func(w *Watcher) {
		w.prefixWatch = true
	}
}

func WatchWithIgnoreErrors() WatcherOption {
	return func(w *Watcher) {
		w.ignoreErrors = true
	}
}

func WatchWithNamespace(namespace *string) WatcherOption {
	return func(w *Watcher) {
		w.namespace = namespace
	}
}

func WatchWithBufSize(bufSize uint32) WatcherOption {
	return func(w *Watcher) {
		w.bufSize = bufSize
	}
}

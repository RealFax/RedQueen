package RedQueen

import (
	"context"
	"github.com/RealFax/RedQueen/locker"
	"github.com/RealFax/RedQueen/store"
	"github.com/hashicorp/raft"
	"time"
)

type LockerBackendWrapper struct {
	store         store.Namespace
	raftApplyFunc func(context.Context, time.Duration, *LogPayload) (raft.ApplyFuture, error)
}

func (w *LockerBackendWrapper) Get(key []byte) (*store.Value, error) {
	return w.store.Get(key)
}

func (w *LockerBackendWrapper) TrySetWithTTL(key, value []byte, ttl uint32) error {
	_, err := w.raftApplyFunc(context.Background(), time.Millisecond*500, &LogPayload{
		Command: TrySetWithTTL,
		TTL:     &ttl,
		Namespace: func() *string {
			ptr := locker.Namespace
			return &ptr
		}(),
		Key:   key,
		Value: value,
	})
	return err
}

func (w *LockerBackendWrapper) Del(key []byte) error {
	_, err := w.raftApplyFunc(context.Background(), time.Millisecond*1000, &LogPayload{
		Command: Del,
		Namespace: func() *string {
			ptr := locker.Namespace
			return &ptr
		}(),
		Key: key,
	})
	return err
}

func (w *LockerBackendWrapper) Watch(key []byte) (store.WatcherNotify, error) {
	return w.store.Watch(key)
}

func NewLockerBackend(
	ns store.Namespace,
	raftApplyFunc func(context.Context, time.Duration, *LogPayload) (raft.ApplyFuture, error),
) locker.Backend {
	return &LockerBackendWrapper{
		store:         ns,
		raftApplyFunc: raftApplyFunc,
	}
}

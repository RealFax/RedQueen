package red

import (
	"context"
	"errors"
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/locker"
	"github.com/RealFax/RedQueen/store"
	"time"
)

type LockerBackendWrapper struct {
	store store.Namespace
	apply func(context.Context, *serverpb.RaftLogPayload, time.Duration) error
}

func (w LockerBackendWrapper) Get(key []byte) (*store.Value, error) {
	return w.store.Get(key)
}

func (w LockerBackendWrapper) TrySetWithTTL(key, value []byte, ttl uint32) error {
	if ttl == 0 {
		return errors.New("race deadlock")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return w.apply(ctx, &serverpb.RaftLogPayload{
		Command: serverpb.RaftLogCommand_TrySetWithTTL,
		Key:     key,
		Value:   value,
		Ttl:     &ttl,
		Namespace: func() *string {
			ptr := locker.Namespace
			return &ptr
		}(),
	}, time.Millisecond*500)
}

func (w LockerBackendWrapper) Del(key []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return w.apply(ctx, &serverpb.RaftLogPayload{
		Command: serverpb.RaftLogCommand_Del,
		Key:     key,
		Namespace: func() *string {
			ptr := locker.Namespace
			return &ptr
		}(),
	}, time.Millisecond*500)
}

func (w LockerBackendWrapper) Watch(key []byte) (store.WatcherNotify, error) {
	return w.store.Watch(key)
}

func NewLockerBackend(
	ns store.Namespace,
	raftApplyFunc func(context.Context, *serverpb.RaftLogPayload, time.Duration) error,
) locker.Backend {
	return &LockerBackendWrapper{
		store: ns,
		apply: raftApplyFunc,
	}
}

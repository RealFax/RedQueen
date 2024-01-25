package dlocker

import (
	"context"
	"errors"
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/internal/rqd/store"
	"time"
)

const (
	Namespace string = "_Locker"
)

type Backend interface {
	Get(key []byte) (*store.Value, error)
	TrySetWithTTL(key, value []byte, ttl uint32) error
	Del(key []byte) error
	Watch(key []byte) (notify store.Watcher, err error)
}

type LockerBackend struct {
	actions store.Actions
	apply   func(context.Context, *serverpb.RaftLogPayload, time.Duration) error
}

func (w LockerBackend) Get(key []byte) (*store.Value, error) {
	return w.actions.Get(key)
}

func (w LockerBackend) TrySetWithTTL(key, value []byte, ttl uint32) error {
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
			ptr := Namespace
			return &ptr
		}(),
	}, time.Millisecond*500)
}

func (w LockerBackend) Del(key []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return w.apply(ctx, &serverpb.RaftLogPayload{
		Command: serverpb.RaftLogCommand_Del,
		Key:     key,
		Namespace: func() *string {
			ptr := Namespace
			return &ptr
		}(),
	}, time.Millisecond*500)
}

func (w LockerBackend) Watch(key []byte) (store.Watcher, error) {
	return w.actions.Watch(key)
}

func NewLockerBackend(
	s store.Store,
	raftApplyFunc func(context.Context, *serverpb.RaftLogPayload, time.Duration) error,
) (Backend, error) {
	current, err := s.Swap(Namespace)
	if err != nil {
		return nil, err
	}
	return &LockerBackend{
		actions: current,
		apply:   raftApplyFunc,
	}, nil
}

package store

import (
	"context"
	"io"
	"time"
)

type Value struct {
	Timestamp uint64
	TTL       uint32
	Key       []byte
	Data      []byte
}

func (v *Value) GetTimestampAsTime() time.Time {
	return time.Unix(int64(v.Timestamp), 0)
}

func (v *Value) GetTTLAsDuration() time.Duration {
	return time.Second * time.Duration(v.TTL)
}

type WatchValue struct {
	Seq       uint64
	Timestamp int64
	TTL       uint32
	Key       []byte
	// Value can be nil pointer, if Value is nil pointer then that the Value is deleted
	Value *[]byte
}

func (v *WatchValue) Deleted() bool {
	return v.Value == nil
}

type Watcher interface {
	Notify() chan *WatchValue
	Close() error
}

type Actions interface {
	Current() string
	Get(key []byte) (value *Value, err error)
	PrefixSearchScan(prefix []byte, reg string, offset, limit int) ([]*Value, error)
	PrefixScan(prefix []byte, offset, limit int) ([]*Value, error)
	SetWithTTL(key, value []byte, ttl uint32) error
	TrySetWithTTL(key, value []byte, ttl uint32) error
	Set(key, value []byte) error
	// TrySet try to set a key-value, returns an error if the key already exists
	TrySet(key, value []byte) error
	Del(key []byte) error
	Watch(key []byte) (notify Watcher, err error)
	WatchPrefix(prefix []byte) Watcher
}

type Store interface {
	Actions
	Swap(namespace string) (Actions, error)
	Close() error
	// Snapshot should be in tar & gzip format
	Snapshot() (io.Reader, error)
	Break(context.Context) error
	Restore(src io.Reader) (err error)
}

package store

import "io"

type Value struct {
	Data []byte
}

func NewValue(data []byte) *Value {
	return &Value{Data: data}
}

type WatchValue struct {
	Seq       uint64
	Timestamp int64
	// Data can be nil pointer, if Data is nil pointer then that the Data is deleted
	Data *[]byte
}

func (v *WatchValue) Deleted() bool {
	return v.Data == nil
}

type WatcherNotify interface {
	Notify() chan *WatchValue
	Close() error
}

type Base interface {
	GetNamespace() string
	Get(key []byte) (value *Value, err error)
	SetWithTTL(key, value []byte, ttl uint32) error
	TrySetWithTTL(key, value []byte, ttl uint32) error
	Set(key, value []byte) error
	// TrySet try to set a key-value, returns an error if the key already exists
	TrySet(key, value []byte) error
	Del(key []byte) error

	Watch(key []byte) (notify WatcherNotify, err error)
}

type Namespace Base

type Store interface {
	Base
	Namespace(namespace string) (Namespace, error)
	Close() error
	// Snapshot should be in tar & gzip format
	Snapshot() (io.Reader, error)
}

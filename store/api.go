package store

type Value struct {
	Data []byte
}

func NewValue(data []byte) *Value {
	return &Value{Data: data}
}

type Base interface {
	GetNamespace() string
	Get(key []byte) (value *Value, err error)
	SetWithTTL(key, value []byte, ttl uint32) error
	SetEXWithTTL(key, value []byte, ttl uint32) error
	Set(key, value []byte) error
	SetEX(key, value []byte) error
	Del(key []byte) error
}

type Namespace Base

type Store interface {
	Base
	Namespace(namespace string) (Namespace, error)
	Close() error
}

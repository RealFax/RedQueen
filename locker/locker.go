package locker

import (
	"github.com/pkg/errors"

	"github.com/RealFax/RedQueen/store"
)

const (
	Namespace string = "_Locker"
)

var (
	ErrStatusBusy = errors.New("status busy")
)

type Backend interface {
	Get(key []byte) (*store.Value, error)
	TrySetWithTTL(key, value []byte, ttl uint32) error
	Del(key []byte) error
	Watch(key []byte) (notify store.WatcherNotify, err error)
}

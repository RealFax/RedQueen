package store

import (
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

const (
	LockDeadline = time.Minute * 10
)

type Lock struct {
	store    Namespace
	Deadline time.Duration
	UUID     string
}

func (l *Lock) Lock() error {
	if err := l.store.SetEXWithTTL([]byte(l.UUID), []byte{0x00}, uint32(l.Deadline)); err != nil {
		if err == ErrKeyAlreadyExists {
			return errors.New("lock status busy")
		}
		return err
	}
	return nil
}

func (l *Lock) Unlock() error {
	val, err := l.store.Get([]byte(l.UUID))
	if len(val.Data) == 0 || err == ErrKeyNotFound {
		return errors.New("lock not found")
	}
	if err = l.store.Del([]byte(l.UUID)); err != nil {
		return err
	}
	return nil
}

func NewLock(s Store) (*Lock, error) {
	ns, err := s.Namespace("locker")
	if err != nil {
		return nil, err
	}
	return &Lock{
		store:    ns,
		UUID:     uuid.New().String(),
		Deadline: LockDeadline,
	}, nil
}

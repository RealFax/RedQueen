package dlocker

import (
	"errors"
	"github.com/RealFax/RedQueen/internal/rqd/store"
	"github.com/RealFax/RedQueen/pkg/hack"
	"time"
)

func MutexLock(lockID string, ttl int32, backend Backend) error {
	if err := backend.TrySetWithTTL(
		hack.String2Bytes(lockID),
		[]byte{},
		func() uint32 {
			if ttl < 0 {
				return 0
			}
			return uint32(ttl)
		}(),
	); err != nil {
		if errors.Is(err, store.ErrKeyAlreadyExists) {
			return ErrStatusBusy
		}
		return err
	}
	return nil
}

func MutexUnlock(lockID string, backend Backend) error {
	if _, err := backend.Get(hack.String2Bytes(lockID)); errors.Is(err, store.ErrKeyNotFound) {
		return ErrStatusBusy
	}
	return backend.Del(hack.String2Bytes(lockID))
}

func MutexTryLock(lockID string, ttl int32, deadline int64, backend Backend) bool {
	if MutexLock(lockID, ttl, backend) == nil {
		return true
	}

	notify, err := backend.Watch(hack.String2Bytes(lockID))
	if err != nil {
		return false
	}
	defer notify.Close()

	ticker := time.NewTicker(time.Duration(time.Now().UnixNano()-deadline) * time.Second)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		return false
	case value := <-notify.Notify():
		if !value.Deleted() {
			return false
		}
		break
	}

	if MutexLock(lockID, ttl, backend) != nil {
		return false
	}

	return true
}

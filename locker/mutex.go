package locker

import (
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/utils"
	"time"
)

func MutexLock(lockID string, ttl int64, backend Backend) error {
	if err := backend.TrySetWithTTL(utils.String2Bytes(lockID), []byte{}, uint32(time.Duration(ttl).Seconds())); err != nil {
		if err == store.ErrKeyAlreadyExists {
			return ErrStatusBusy
		}
		return err
	}
	return nil
}

func MutexUnlock(lockID string, backend Backend) error {
	val, err := backend.Get(utils.String2Bytes(lockID))
	if len(val.Data) == 0 || err == store.ErrKeyNotFound {
		return ErrStatusBusy
	}
	return backend.Del(utils.String2Bytes(lockID))
}

func MutexTryLock(lockID string, ttl int64, deadline int64, backend Backend) bool {
	if MutexLock(lockID, ttl, backend) == nil {
		return true
	}

	notify, err := backend.Watch(utils.String2Bytes(lockID))
	if err != nil {
		return false
	}
	defer notify.Close()

	ticker := time.NewTicker(time.Duration(deadline))
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

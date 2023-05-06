package locker

import (
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/utils"
	"github.com/google/uuid"
	"time"
)

type Mutex struct {
	store  Backend
	TTL    time.Duration
	LockID string
}

type MutexOption func(*Mutex)

func (l *Mutex) Type() string {
	return "mutex"
}

func (l *Mutex) Lock() error {
	if err := l.store.TrySetWithTTL(utils.String2Bytes(l.LockID), []byte{}, uint32(l.TTL.Seconds())); err != nil {
		if err == store.ErrKeyAlreadyExists {
			return ErrStatusBusy
		}
		return err
	}
	return nil
}

func (l *Mutex) Unlock() error {
	val, err := l.store.Get(utils.String2Bytes(l.LockID))
	if len(val.Data) == 0 || err == store.ErrKeyNotFound {
		return ErrStatusBusy
	}
	return l.store.Del(utils.String2Bytes(l.LockID))
}

//func (l *Mutex) TryLock(deadline time.Duration) error {
//	ctx, cancel := context.WithTimeout(context.Background(), deadline)
//	for {
//		select {
//		case <-ctx.Done():
//			cancel()
//			return ctx.Err()
//		default:
//			// success locked
//			if err := l.Lock(); err == nil {
//				cancel()
//				return nil
//			}
//			// random time slice
//			n, _ := rand.Int(rand.Reader, big.NewInt(1000))
//			time.Sleep(time.Microsecond * time.Duration(n.Int64()))
//		}
//	}
//}

// TryLock tries to lock m and reports whether it succeeded.
//
// Note that if the mutex is not released before reaching the deadline
// it will wait until it is released, and it maybe not succeed
func (l *Mutex) TryLock(deadline time.Duration) bool {
	if l.Lock() == nil {
		return true
	}

	notify, err := l.store.Watch(utils.String2Bytes(l.LockID))
	if err != nil {
		return false
	}
	defer notify.Close()

	ticker := time.NewTicker(deadline)
	defer ticker.Stop()

	select {
	case <-ticker.C:
		return false
	// There may be one or more clients waiting for the mutex,
	// but we are now watched for this lock to be removed,
	// and can try to compete for that mutex after the lock is deleted
	case value := <-notify.Notify():
		if !value.Deleted() {
			return false
		}
		break
	}

	if l.Lock() != nil {
		return false
	}

	return true
}

func NewMutex(s Backend, options ...MutexOption) (*Mutex, error) {
	mutex := &Mutex{
		store:  s,
		LockID: uuid.New().String(),
		TTL:    TTL,
	}

	for _, option := range options {
		option(mutex)
	}

	return mutex, nil
}

func MutexWithTTL(deadline time.Duration) func(mutex *Mutex) {
	return func(mutex *Mutex) {
		mutex.TTL = deadline
	}
}

func MutexWithCustomID(id string) func(mutex *Mutex) {
	return func(mutex *Mutex) {
		mutex.LockID = id
	}
}

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

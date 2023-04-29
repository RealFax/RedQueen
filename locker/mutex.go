package locker

import (
	"time"

	"github.com/google/uuid"

	"github.com/RealFax/RedQueen/store"
)

type Mutex struct {
	store    store.Namespace
	Deadline time.Duration
	UUID     string
}

func (l *Mutex) Lock() error {
	if err := l.store.SetEXWithTTL([]byte(l.UUID), []byte{0x00}, uint32(l.Deadline)); err != nil {
		if err == store.ErrKeyAlreadyExists {
			return ErrStatusBusy
		}
		return err
	}
	return nil
}

func (l *Mutex) Unlock() error {
	val, err := l.store.Get([]byte(l.UUID))
	if len(val.Data) == 0 || err == store.ErrKeyNotFound {
		return ErrStatusBusy
	}
	if err = l.store.Del([]byte(l.UUID)); err != nil {
		return err
	}
	return nil
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

func (l *Mutex) TryLock() bool {
	notify, err := l.store.Watch([]byte(l.UUID))
	if err != nil {
		return false
	}
	defer notify.Close()

	// There may be one or more clients waiting for the mutex,
	// but we are now watched for this lock to be removed,
	// and can try to compete for that mutex after the lock is deleted
	value := <-notify.Notify()
	if !value.Deleted() {
		return false
	}

	if err = l.Lock(); err != nil {
		return false
	}

	return true
}

func NewMutex(s store.Store) (*Mutex, error) {
	ns, err := s.Namespace(Namespace)
	if err != nil {
		return nil, err
	}
	return &Mutex{
		store:    ns,
		UUID:     uuid.New().String(),
		Deadline: Deadline,
	}, nil
}

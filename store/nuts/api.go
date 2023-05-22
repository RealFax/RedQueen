package nuts

import (
	"bytes"
	"context"
	"github.com/RealFax/RedQueen/store"
	"github.com/nutsdb/nutsdb"
	"github.com/pkg/errors"
	"io"
	"sync"
	"sync/atomic"
)

func (s *storeAPI) _namespace(namespace string) (*storeAPI, error) {
	if atomic.LoadUint32(s.state) == StateBreak {
		return nil, ErrStateBreak
	}
	if namespace == s.namespace {
		return s, nil
		// return nil, errors.New("conflicts with the current namespace")
	}
	return &storeAPI{
		state:        s.state,
		db:           s.db,
		watcher:      s.watcher,
		watcherChild: s.watcher.Namespace(namespace),
		mu:           sync.Mutex{},
		namespace:    namespace,
	}, nil
}

func (s *storeAPI) State() uint32 {
	return atomic.LoadUint32(s.state)
}

func (s *storeAPI) DB() (*nutsdb.DB, error) {
	if atomic.LoadUint32(s.state) == StateBreak {
		return nil, ErrStateBreak
	}
	return s.db, nil
}

func (s *storeAPI) Transaction(writable bool, fn func(tx *nutsdb.Tx) error) error {
	db, err := s.DB()
	if err != nil {
		return err
	}

	var tx *nutsdb.Tx
	if tx, err = db.Begin(writable); err != nil {
		return err
	}

	if err = fn(tx); err != nil {
		tx.Rollback()
		return err
	} else {
		if err = tx.Commit(); err != nil {
			tx.Rollback()
			return err
		}
	}
	return nil
}

func (s *storeAPI) Get(key []byte) (*store.Value, error) {
	db, err := s.DB()
	if err != nil {
		return nil, err
	}
	val := store.NewValue(nil)
	return val, db.View(func(tx *nutsdb.Tx) error {
		entry, gErr := tx.Get(s.namespace, key)
		if gErr != nil {
			if gErr == nutsdb.ErrKeyNotFound {
				return store.ErrKeyNotFound
			}
			return gErr
		}
		val.Data = entry.Value
		return nil
	})
}

func (s *storeAPI) SetWithTTL(key, value []byte, ttl uint32) error {
	return s.Transaction(true, func(tx *nutsdb.Tx) error {
		if err := tx.Put(s.namespace, key, value, ttl); err != nil {
			return err
		}
		s.watcherChild.Update(key, value)
		return nil
	})
}

func (s *storeAPI) Set(key, value []byte) error {
	return s.SetWithTTL(key, value, 0)
}

func (s *storeAPI) TrySetWithTTL(key, value []byte, ttl uint32) error {
	db, err := s.DB()
	if err != nil {
		return err
	}
	return db.Update(func(tx *nutsdb.Tx) error {
		_, err = tx.Get(s.namespace, key)
		if err == nil {
			return store.ErrKeyAlreadyExists
		}

		if err = tx.Put(s.namespace, key, value, ttl); err != nil {
			return err
		}

		// notify watcher key-value update
		s.watcherChild.Update(key, value)
		return nil
	})
}

func (s *storeAPI) TrySet(key, value []byte) error {
	return s.TrySetWithTTL(key, value, 0)
}

func (s *storeAPI) Del(key []byte) error {
	db, err := s.DB()
	if err != nil {
		return err
	}
	return db.Update(func(tx *nutsdb.Tx) error {
		if err = tx.Delete(s.namespace, key); err != nil {
			return err
		}
		s.watcherChild.Update(key, nil)
		return nil
	})
}

func (s *storeAPI) Watch(key []byte) (store.WatcherNotify, error) {
	if strictMode.Load() {
		db, err := s.DB()
		if err != nil {
			return nil, err
		}
		// check watch key does it exist
		if err = db.View(func(tx *nutsdb.Tx) error {
			_, err = tx.Get(s.namespace, key)
			return err
		}); err != nil {
			return nil, errors.Wrap(err, "strict")
		}
	}
	return s.watcherChild.Watch(key), nil
}

func (s *storeAPI) GetNamespace() string {
	return s.namespace
}

func (s *storeAPI) Namespace(namespace string) (store.Namespace, error) {
	return s._namespace(namespace)
}

func (s *storeAPI) Close() error {
	db, err := s.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func (s *storeAPI) Snapshot() (io.Reader, error) {
	db, err := s.DB()
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err = db.Merge(); err != nil {
		return nil, errors.New("fatal snapshot")
	}
	buf := &bytes.Buffer{}
	return buf, db.View(func(tx *nutsdb.Tx) error {
		return db.BackupTarGZ(buf)
	})
}

func (s *storeAPI) Break(ctx context.Context) error {
	if atomic.LoadUint32(s.state) == StateBreak {
		return ErrStateBreak
	}

	// switch to state: break
	s.mu.Lock()
	atomic.StoreUint32(s.state, StateBreak)

	go func() {
		select {
		case <-ctx.Done():
			s.mu.Unlock()
			atomic.StoreUint32(s.state, StateOk)
		}
	}()

	return nil
}

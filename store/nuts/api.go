package nuts

import (
	"bytes"
	"context"
	"github.com/RealFax/RedQueen/store"
	"github.com/nutsdb/nutsdb"
	"github.com/pkg/errors"
	"io"
	"os"
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
	val := store.NewValue(nil)
	return val, s.Transaction(false, func(tx *nutsdb.Tx) error {
		entry, err := tx.Get(s.namespace, key)
		if err != nil {
			if err == nutsdb.ErrKeyNotFound {
				return store.ErrKeyNotFound
			}
			return err
		}
		val.Timestamp = entry.Meta.Timestamp
		val.TTL = entry.Meta.TTL
		val.Key = entry.Key
		val.Data = entry.Value
		return nil
	})
}

func (s *storeAPI) PrefixSearchScan(prefix []byte, reg string, offset, limit int) ([]*store.Value, error) {
	val := make([]*store.Value, 0, limit-offset)
	return val, s.Transaction(false, func(tx *nutsdb.Tx) error {
		var (
			err     error
			entries nutsdb.Entries
		)

		if reg != "" {
			entries, _, err = tx.PrefixSearchScan(s.namespace, prefix, reg, offset, limit)
		} else {
			entries, _, err = tx.PrefixScan(s.namespace, prefix, offset, limit)
		}
		if err != nil {
			return err
		}

		for _, entry := range entries {
			val = append(val, &store.Value{
				Timestamp: entry.Meta.Timestamp,
				TTL:       entry.Meta.TTL,
				Key:       entry.Key,
				Data:      entry.Value,
			})
		}
		return nil
	})
}

func (s *storeAPI) PrefixScan(prefix []byte, offset, limit int) ([]*store.Value, error) {
	return s.PrefixSearchScan(prefix, "", offset, limit)
}

func (s *storeAPI) SetWithTTL(key, value []byte, ttl uint32) error {
	return s.Transaction(true, func(tx *nutsdb.Tx) error {
		if err := tx.Put(s.namespace, key, value, ttl); err != nil {
			return err
		}
		// notify watcher key-value update
		s.watcherChild.Update(key, value)
		return nil
	})
}

func (s *storeAPI) Set(key, value []byte) error {
	return s.SetWithTTL(key, value, 0)
}

func (s *storeAPI) TrySetWithTTL(key, value []byte, ttl uint32) error {
	return s.Transaction(true, func(tx *nutsdb.Tx) error {
		_, err := tx.Get(s.namespace, key)
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
	return s.Transaction(true, func(tx *nutsdb.Tx) error {
		if err := tx.Delete(s.namespace, key); err != nil {
			return err
		}
		s.watcherChild.Update(key, nil)
		return nil
	})
}

func (s *storeAPI) Watch(key []byte) (store.WatcherNotify, error) {
	if strictMode.Load() {
		// check watch key does it exist
		if err := s.Transaction(false, func(tx *nutsdb.Tx) error {
			_, err := tx.Get(s.namespace, key)
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

func (s *storeAPI) GetWatch() store.WatcherMetadata {
	return s.watcher.Get()
}

func (s *storeAPI) Namespace(namespace string) (store.Namespace, error) {
	n, err := s._namespace(namespace)
	if err != nil {
		return nil, err
	}
	_ = n.SetWithTTL(initBucketKey, nil, 4) // init bucket
	return n, nil
}

func (s *storeAPI) Close() error {
	db, err := s.DB()
	if err != nil {
		return err
	}
	s.Break(context.Background())
	return db.Close()
}

func (s *storeAPI) Snapshot() (io.Reader, error) {
	// get db session first
	db, err := s.DB()
	if err != nil {
		return nil, errors.Wrap(err, "fail snapshot, get db error")
	}

	// break db
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err = s.Break(ctx); err != nil {
		return nil, errors.Wrap(err, "fail snapshot, break error")
	}

	db.Merge()
	//if err = db.Merge(); err != nil {
	//	return nil, errors.Wrap(err, "fail snapshot, merge error")
	//}

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

func (s *storeAPI) Restore(src io.Reader) (err error) {
	ctx, cancel := context.WithCancel(context.Background())

	if err = s.Break(ctx); err != nil {
		goto CancelBreak
	}

	// close nuts db
	s.db.Close()

	// clear new db files
	if err = os.RemoveAll(s.dataDir); err != nil {
		err = errors.Wrap(err, "clean files error")
		goto CancelBreak
	}

	// restore old db files
	if err = BackupReader(s.dataDir, src); err != nil {
		err = errors.Wrap(err, "")
		goto CancelBreak
	}

	// reopen nuts db
	if s.db, err = nutsdb.Open(nutsdb.DefaultOptions, s.dbOptions...); err != nil {
		goto CancelBreak
	}

	goto CancelBreak

CancelBreak:
	cancel()
	return
}

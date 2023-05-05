package nuts

import (
	"bytes"
	"github.com/RealFax/RedQueen/store"
	"github.com/nutsdb/nutsdb"
	"github.com/pkg/errors"
	"io"
	"sync"
)

func (s *storeAPI) _namespace(namespace string) (*storeAPI, error) {
	if namespace == s.namespace {
		return s, nil
		// return nil, errors.New("conflicts with the current namespace")
	}
	return &storeAPI{
		db:           s.db,
		watcher:      s.watcher,
		watcherChild: s.watcher.Namespace(namespace),
		mu:           sync.Mutex{},
		namespace:    namespace,
	}, nil
}

func (s *storeAPI) Close() error {
	return s.db.Close()
}

func (s *storeAPI) Get(key []byte) (*store.Value, error) {
	val := store.NewValue(nil)
	return val, s.db.View(func(tx *nutsdb.Tx) error {
		entry, err := tx.Get(s.namespace, key)
		if err != nil {
			if err == nutsdb.ErrKeyNotFound {
				return store.ErrKeyNotFound
			}
			return err
		}
		val.Data = entry.Value
		return nil
	})
}

func (s *storeAPI) SetWithTTL(key, value []byte, ttl uint32) error {
	return s.db.Update(func(tx *nutsdb.Tx) error {
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
	return s.db.Update(func(tx *nutsdb.Tx) error {
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
	return s.db.Update(func(tx *nutsdb.Tx) error {
		if err := tx.Delete(s.namespace, key); err != nil {
			return err
		}
		s.watcherChild.Update(key, nil)
		return nil
	})
}

func (s *storeAPI) Watch(key []byte) (store.WatcherNotify, error) {
	if strictMode {
		// check watch key does it exist
		if err := s.db.View(func(tx *nutsdb.Tx) error {
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

func (s *storeAPI) Snapshot() (io.Reader, error) {
	if err := s.db.Merge(); err != nil {
		return nil, errors.New("fatal snapshot")
	}
	buf := &bytes.Buffer{}
	return buf, s.db.View(func(tx *nutsdb.Tx) error {
		return s.db.BackupTarGZ(buf)
	})
}

func (s *storeAPI) Namespace(namespace string) (store.Namespace, error) {
	return s._namespace(namespace)
}

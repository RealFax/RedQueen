package nuts

import (
	"github.com/RealFax/RedQueen/store"
	"github.com/nutsdb/nutsdb"
	"github.com/pkg/errors"
	"sync"
)

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
		return tx.Put(s.namespace, key, value, ttl)
	})
}

func (s *storeAPI) Set(key, value []byte) error {
	return s.SetWithTTL(key, value, 0)
}

func (s *storeAPI) SetEXWithTTL(key, value []byte, ttl uint32) error {
	return s.db.Update(func(tx *nutsdb.Tx) error {
		_, err := tx.Get(s.namespace, key)
		if err == nil {
			return store.ErrKeyAlreadyExists
		}
		return tx.Put(s.namespace, key, value, ttl)
	})
}

func (s *storeAPI) SetEX(key, value []byte) error {
	return s.SetEXWithTTL(key, value, 0)
}

func (s *storeAPI) Del(key []byte) error {
	return s.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Delete(s.namespace, key)
	})
}

func (s *storeAPI) GetNamespace() string {
	return s.namespace
}

func (s *storeAPI) Namespace(namespace string) (store.Namespace, error) {
	if namespace == s.namespace {
		return nil, errors.New("conflicts with the current namespace")
	}
	return &storeAPI{
		db:        s.db,
		mu:        sync.Mutex{},
		namespace: namespace,
	}, nil
}

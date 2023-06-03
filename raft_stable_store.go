package red

import (
	"encoding/binary"
	"github.com/RealFax/RedQueen/store"
)

type StableStore struct {
	store store.Namespace
}

func (s *StableStore) Set(key []byte, val []byte) error {
	return s.store.Set(key, val)
}

// Get returns the value for key, or an empty byte slice if key was not found.
func (s *StableStore) Get(key []byte) ([]byte, error) {
	return store.UnwrapGet(s.store.Get(key))
}

func (s *StableStore) SetUint64(key []byte, val uint64) error {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, val)
	return s.store.Set(key, buf)
}

// GetUint64 returns the uint64 value for key, or 0 if key was not found.
func (s *StableStore) GetUint64(key []byte) (uint64, error) {
	val, err := store.UnwrapGet(s.store.Get(key))
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(val), nil
}

func NewStableStore(s store.Store) (*StableStore, error) {
	namespace, err := s.Namespace("_RaftStableStore")
	if err != nil {
		return nil, err
	}
	return &StableStore{store: namespace}, nil
}

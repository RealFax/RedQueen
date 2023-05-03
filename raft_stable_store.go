package RedQueen

import (
	"bytes"
	"encoding/binary"
	"github.com/RealFax/RedQueen/store"
)

type StableStore struct {
	Store store.Store
}

func (s *StableStore) Set(key []byte, val []byte) error {
	return s.Store.Set(key, val)
}

// Get returns the value for key, or an empty byte slice if key was not found.
func (s *StableStore) Get(key []byte) ([]byte, error) {
	val, err := s.Store.Get(key)
	if err != nil {
		return nil, err
	}
	return val.Data, nil
}

func (s *StableStore) SetUint64(key []byte, val uint64) error {
	buf := &bytes.Buffer{}
	binary.Write(buf, binary.LittleEndian, val)
	return s.Store.Set(key, buf.Bytes())
}

// GetUint64 returns the uint64 value for key, or 0 if key was not found.
func (s *StableStore) GetUint64(key []byte) (uint64, error) {
	data, err := s.Store.Get(key)
	if err != nil {
		return 0, err
	}

	var (
		buf = bytes.NewReader(data.Data)
		n   uint64
	)
	if err = binary.Read(buf, binary.LittleEndian, &n); err != nil {
		return 0, err
	}
	return n, nil
}

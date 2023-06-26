package collapsar

import (
	"encoding/gob"
	"fmt"
	"io"
	"sync/atomic"
)

type reader struct {
	size, offset int32
	entries      [][]byte
}

func (r *reader) Next() ([]byte, error) {
	if atomic.LoadInt32(&r.size) == atomic.LoadInt32(&r.offset) {
		return nil, io.EOF
	}

	return r.entries[atomic.AddInt32(&r.offset, 1)-1], nil
}

func NewReader(r io.Reader) (Reader, error) {
	var bin Binary
	if err := gob.NewDecoder(r).Decode(&bin); err != nil {
		fmt.Println(bin)
		return nil, err
	}
	return &reader{
		size:    bin.Size,
		entries: bin.Entries,
	}, nil
}

package collapsar

import "io"

type Writer interface {
	Encode(w io.Writer) error
	Add(entry []byte) error
	Wait()
}

type Reader interface {
	Next() ([]byte, error)
}

type Binary struct {
	Size    int32
	Entries [][]byte
}

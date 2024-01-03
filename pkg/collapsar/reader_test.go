package collapsar_test

import (
	"bytes"
	collapsar2 "github.com/RealFax/RedQueen/pkg/collapsar"
	"io"
	"os"
	"testing"
)

func TestNewReader(t *testing.T) {
	collapsar2.NewReader(os.Stdin)
}

func TestReader_Next(t *testing.T) {
	w := collapsar2.NewWriter(4)
	w.Add([]byte{1})
	w.Add([]byte{2})
	w.Add([]byte{3})
	w.Add([]byte{4})

	buf := &bytes.Buffer{}
	if err := w.Encode(buf); err != nil {
		t.Fatal(err)
	}

	r, err := collapsar2.NewReader(buf)
	if err != nil {
		t.Fatal(err)
	}

	var p []byte
	for {
		if p, err = r.Next(); err != nil {
			if err == io.EOF {
				return
			}
			t.Fatal(err)
		}
		t.Log(p)
	}

}

package collapsar_test

import (
	"bytes"
	"github.com/RealFax/RedQueen/pkg/collapsar"
	"testing"
)

func TestWriter_Encode(t *testing.T) {
	w := collapsar.NewWriter(1)
	if err := w.Add([]byte{}); err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	if err := w.Encode(buf); err != nil {
		t.Fatal(err)
	}

	t.Logf("[+] Encoding size: %d", buf.Len())
}

func TestWriter_Add(t *testing.T) {
	w := collapsar.NewWriter(1)
	if err := w.Add([]byte{}); err != nil {
		t.Fatal(err)
	}
}

func TestWriter_Wait(t *testing.T) {
	w := collapsar.NewWriter(4)

	go func() {
		for i := 0; i < 4; i++ {
			w.Add([]byte{byte(i)})
			t.Logf("[+] Add: %d", i+1)
		}
	}()

	w.Wait()
	t.Log("[+] Done")
}

func TestWriter_Close(t *testing.T) {
	w := collapsar.NewWriter(2)
	if err := w.Add([]byte{1}); err != nil {
		t.Fatal("unexpected error:", err)
	}
	w.Close()
	if err := w.Add([]byte{1}); err != nil {
		t.Log("expected error:", err)
	}
}

func BenchmarkWriter_Add(b *testing.B) {
	w := collapsar.NewWriter(int32(b.N))
	for i := 0; i < b.N; i++ {
		w.Add([]byte{byte(i)})
	}
}

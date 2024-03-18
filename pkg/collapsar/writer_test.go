package collapsar_test

import (
	"bytes"
	"github.com/RealFax/RedQueen/pkg/collapsar"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWriter_Encode(t *testing.T) {
	w := collapsar.NewWriter(1)
	assert.NoError(t, w.Add([]byte{}))

	buf := &bytes.Buffer{}
	assert.NoError(t, w.Encode(buf))

	t.Logf("[+] Encoding size: %d", buf.Len())
}

func TestWriter_Add(t *testing.T) {
	w := collapsar.NewWriter(1)
	assert.NoError(t, w.Add([]byte{}))
}

func TestWriter_Wait(t *testing.T) {
	w := collapsar.NewWriter(4)

	go func() {
		for i := 0; i < 4; i++ {
			assert.NoError(t, w.Add([]byte{byte(i)}))
		}
	}()

	w.Wait()
}

func TestWriter_Close(t *testing.T) {
	w := collapsar.NewWriter(2)
	assert.NoError(t, w.Add([]byte{1}))
	assert.NoError(t, w.Close())
	assert.Error(t, w.Add([]byte{1}))
}

func BenchmarkWriter_Add(b *testing.B) {
	w := collapsar.NewWriter(int32(b.N))
	for i := 0; i < b.N; i++ {
		assert.NoError(b, w.Add([]byte{byte(i)}))
	}
}

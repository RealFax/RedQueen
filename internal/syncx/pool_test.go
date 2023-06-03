package syncx_test

import (
	"bytes"
	"github.com/RealFax/RedQueen/internal/syncx"
	"testing"
)

var (
	bytesBufferPool = syncx.NewPool[*bytes.Buffer](
		func() *bytes.Buffer { return &bytes.Buffer{} },
		nil,
		func(val *bytes.Buffer) { val.Reset() },
	)
)

func TestPool_Alloc(t *testing.T) {
	buf := bytesBufferPool.Alloc()

	buf.Val().WriteString("test")
	t.Log(buf.Val().String())

	buf.Free()
}

func BenchmarkPool_Free(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bytesBufferPool.Alloc().Free()
	}
}

package nuts_test

import (
	"fmt"
	"github.com/RealFax/RedQueen/internal/rqd/store/nuts"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWatchKey(t *testing.T) {
	watchKey := nuts.WatchKey([]byte("Test"))
	// simple hash collision test
	for i := 0; i < 10000000; i++ {
		assert.NotEqual(t, watchKey, nuts.WatchKey([]byte(fmt.Sprintf("Test%d", i))))
	}
}

func BenchmarkWatchKey(b *testing.B) {
	k := []byte("Test")
	for i := 0; i < b.N; i++ {
		nuts.WatchKey(k)
	}
}

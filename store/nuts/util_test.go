package nuts_test

import (
	"github.com/RealFax/RedQueen/store/nuts"
	"testing"
)

func TestWatchKey(t *testing.T) {
	nuts.WatchKey([]byte("Test"))
}

func BenchmarkWatchKey(b *testing.B) {
	k := []byte("Test")
	for i := 0; i < b.N; i++ {
		nuts.WatchKey(k)
	}
}

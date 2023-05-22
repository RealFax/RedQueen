package nuts_test

import (
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/store/nuts"
	"testing"
)

var db store.Store

func init() {
	var err error
	if db, err = nuts.New(nuts.Config{
		NodeNum: 4,
		Sync:    false,
		DataDir: "/tmp/nuts-db",
	}); err != nil {
		panic(err)
	}

	db.Set([]byte("Hello"), []byte("World"))
}

func BenchmarkStoreAPI_Get(b *testing.B) {
	k := []byte("Hello")
	for i := 0; i < b.N; i++ {
		db.Get(k)
	}
}

func BenchmarkStoreAPI_Set(b *testing.B) {
	k, v := []byte("Hello"), []byte("World")
	for i := 0; i < b.N; i++ {
		db.Set(k, v)
	}
}

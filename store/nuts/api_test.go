package nuts_test

import (
	"context"
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/store/nuts"
	"testing"
	"time"
)

var db store.Store

func init() {
	var err error
	if db, err = nuts.New(nuts.Config{
		NodeNum: 1,
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

func BenchmarkStoreAPI_SetWithTTL(b *testing.B) {
	k, v := []byte("Hello"), []byte("World")
	for i := 0; i < b.N; i++ {
		db.SetWithTTL(k, v, 1)
	}
}

func TestStoreAPI_Break(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	if err := db.Break(ctx); err != nil {
		t.Fatal(err)
	}

	if err := db.Set([]byte("Hello"), []byte("World")); err != nil {
		t.Error(err)
	}

	cancel()

	time.Sleep(time.Millisecond * 100)

	if err := db.Set([]byte("Hello"), []byte("World")); err != nil {
		t.Error(err)
	}

}

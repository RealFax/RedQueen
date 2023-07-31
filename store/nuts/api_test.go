package nuts_test

import (
	"bytes"
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/store/nuts"
)

var (
	db         store.Store
	key, value = []byte("Hello"), []byte("World")
)

func init() {
	os.RemoveAll("/tmp/nuts-db")
	var err error
	if db, err = nuts.New(nuts.Config{
		NodeNum: 1,
		Sync:    false,
		DataDir: "/tmp/nuts-db",
		RWMode:  nuts.MMap,
	}); err != nil {
		panic(err)
	}

	db.Set(key, value)
}

func getWithPrint(t *testing.T, key []byte, passErr bool) {
	val, err := db.Get(key)
	if err != nil {
		if !passErr {
			t.Fatal(err)
		}
		t.Log("PassError:", err)
		return
	}
	t.Logf("Value: %s, Timestamp: %d, TTL: %d", val.Data, val.Timestamp, val.TTL)
}

func TestStoreAPI_Get(t *testing.T) {
	getWithPrint(t, key, false)
}

func TestStoreAPI_PrefixSearchScan(t *testing.T) {
	for off := 0; off < 10; off++ {
		db.Set([]byte("user_"+strconv.Itoa(off)+"_state"), []byte(strconv.Itoa(off)))
	}

	result, err := db.PrefixSearchScan([]byte("user_"), "[^0-9_]", 0, 10)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range result {
		t.Logf("Value: %s, Timestamp: %d, TTL: %d", v.Data, v.Timestamp, v.TTL)
	}
}

func TestStoreAPI_PrefixScan(t *testing.T) {
	for off := 0; off < 100; off++ {
		db.Set([]byte("user_"+strconv.Itoa(off)+"_state"), []byte(strconv.Itoa(off)))
	}

	result, err := db.PrefixScan([]byte("user_"), 0, 50)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Result size: ", len(result))
}

func TestStoreAPI_SetWithTTL(t *testing.T) {
	if err := db.SetWithTTL([]byte("SetWithTTLKey"), []byte("SetWithTTlValue"), 16); err != nil {
		t.Fatal(err)
	}
	getWithPrint(t, []byte("SetWithTTLKey"), false)
}

func TestStoreAPI_Set(t *testing.T) {
	if err := db.Set([]byte("SetKey"), []byte("SetValue")); err != nil {
		t.Fatal(err)
	}
	getWithPrint(t, []byte("SetKey"), false)
}

func TestStoreAPI_TrySetWithTTL(t *testing.T) {
	if err := db.TrySetWithTTL([]byte("TrySetWithTTLKey"), []byte("TrySetWithTTLValue"), 16); err != nil {
		t.Fatal(err)
	}
	getWithPrint(t, []byte("TrySetWithTTLKey"), false)

	if err := db.TrySetWithTTL([]byte("TrySetWithTTLKey"), nil, 16); err != nil {
		t.Log("expected error:", err)
	}
	getWithPrint(t, []byte("TrySetWithTTLKey"), false)
}

func TestStoreAPI_TrySet(t *testing.T) {
	if err := db.TrySet([]byte("TrySetKey"), []byte("TrySetValue")); err != nil {
		t.Fatal(err)
	}
	getWithPrint(t, []byte("TrySetKey"), false)

	if err := db.TrySet([]byte("TrySetKey"), nil); err != nil {
		t.Log("expected error:", err)
	}
	getWithPrint(t, []byte("TrySetKey"), false)
}

func TestStoreAPI_Del(t *testing.T) {
	if err := db.Set([]byte("DelKey"), []byte("DelValue")); err != nil {
		t.Fatal(err)
	}
	getWithPrint(t, []byte("DelKey"), false)

	if err := db.Del([]byte("DelKey")); err != nil {
		t.Fatal(err)
	}
	getWithPrint(t, []byte("DelKey"), true)
}

func TestStoreAPI_Watch(t *testing.T) {
	// watch before set, strict mode must be disabled first
	nuts.DisableStrictMode()

	notify, err := db.Watch(keys[0])
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		t.Log("[+] Started watch")
		for {
			select {
			case val := <-notify.Notify():
				t.Logf("Seq: %d, Timestamp: %d, Data: %s", val.Seq, val.Timestamp, *val.Data)
			case <-ctx.Done():
				t.Log("[+] End watch")
				notify.Close()
				return
			}
		}
	}()

	time.Sleep(time.Second * 1)

	for i := 0; i < 10; i++ {
		if err = db.Set(keys[0], []byte("Hello, Watcher")); err != nil {
			t.Fatal("set failed:", err)
		}
		time.Sleep(time.Millisecond * 300)
	}

	t.Log("[+] waiting watcher...")

	time.Sleep(time.Second * 1)

	cancel()

	time.Sleep(time.Second * 1)
}

func TestStoreAPI_Namespace(t *testing.T) {
	t.Logf("current namespace: %s", db.GetNamespace())
	namespace, err := db.Namespace("NextNamespace")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("current namespace: %s", namespace.GetNamespace())
}

func TestStoreAPI_Snapshot(t *testing.T) {
	snapshot, err := db.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("snapshot size: %d", snapshot.(*bytes.Buffer).Len())
}

func TestStoreAPI_Break(t *testing.T) {
	time.Sleep(time.Millisecond * 50) // waiting quit break state
	ctx, cancel := context.WithCancel(context.Background())
	if err := db.Break(ctx); err != nil {
		cancel()
		t.Fatal(err)
	}
	getWithPrint(t, key, true)
	cancel()                          // cancel break state
	time.Sleep(time.Millisecond * 50) // waiting quit break state
	getWithPrint(t, key, false)
}

func TestStoreAPI_Restore(t *testing.T) {
	snapshot, err := db.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("snapshot size: %d", snapshot.(*bytes.Buffer).Len())
	time.Sleep(time.Millisecond * 50) // waiting quit break state
	if err = db.Restore(snapshot); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond * 50) // waiting quit break state
	getWithPrint(t, key, false)
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

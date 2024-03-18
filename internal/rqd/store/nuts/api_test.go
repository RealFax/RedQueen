package nuts_test

import (
	"context"
	"github.com/RealFax/RedQueen/internal/rqd/store"
	"github.com/RealFax/RedQueen/internal/rqd/store/nuts"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

type TestPair struct {
	Key, Value []byte
}

var (
	dir   string
	db    store.Store
	pair1 = TestPair{
		Key:   []byte("Key1"),
		Value: []byte("Value1"),
	}
	pair2 = TestPair{
		Key:   []byte("Key2"),
		Value: []byte("Value2"),
	}
	pairMatrix = []TestPair{
		{Key: []byte("K0"), Value: []byte("V0")},
		{Key: []byte("K1"), Value: []byte("V1")},
		{Key: []byte("K2"), Value: []byte("V2")},
		{Key: []byte("K3"), Value: []byte("V3")},
		{Key: []byte("K4"), Value: []byte("V4")},
		{Key: []byte("K5"), Value: []byte("V5")},
		{Key: []byte("K6"), Value: []byte("V6")},
		{Key: []byte("K7"), Value: []byte("V7")},
		{Key: []byte("K8"), Value: []byte("V8")},
		{Key: []byte("K9"), Value: []byte("V9")},
	}
)

func reset() {
	var err error
	dir, err = os.MkdirTemp("", "nuts-db")
	if err != nil {
		panic("init testing error, cause:" + err.Error())
	}
	if db, err = nuts.New(nuts.Config{
		NodeNum: 1,
		Sync:    false,
		DataDir: dir,
		RWMode:  nuts.MMap,
	}); err != nil {
		panic(err)
	}

	db.Set(pair1.Key, pair1.Value)
}

func ShouldTimeout(t *testing.T, timeout time.Duration, fc func()) {
	t.Helper()
	fin := make(chan struct{})

	go func() {
		defer close(fin)
		fc()
	}()

	select {
	case <-fin:
		t.Errorf("test case should timeout %s", timeout)
	case <-time.After(timeout):
		t.SkipNow()
		os.Exit(1)
		return
	}
}

func TestDB_Swap(t *testing.T) {
	reset()

	db1, err := db.Swap("bucket1")
	assert.NoError(t, err)
	assert.NotNil(t, db1)
	assert.Equal(t, "bucket1", db1.Current())

	db2, err := db.Swap("bucket2")
	assert.NoError(t, err)
	assert.NotNil(t, db2)
	assert.Equal(t, "bucket2", db2.Current())

	// test to repeatedly Swap the same namespace
	db3, err := db.Swap("bucket1")
	assert.NoError(t, err)
	assert.Equal(t, db3, db1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err = db.Break(ctx) // expect: not error
	assert.NoError(t, err)

	_, err = db.Swap("bucket0") // expect: return error
	assert.Error(t, err)
}

func TestDB_Break(t *testing.T) {
	reset()

	value, err := db.Get(pair1.Key)
	assert.NoError(t, err)
	assert.Equal(t, value.Data, pair1.Value)

	ctx, cancel := context.WithCancel(context.Background())
	err = db.Break(ctx)
	assert.NoError(t, err)

	// expect: timeout...
	ShouldTimeout(t, 1*time.Second, func() {
		value, err = db.Get(pair1.Key)
		assert.Error(t, err)
		assert.NotNil(t, value)
	})

	err = db.Break(ctx) // expect: return error
	assert.Error(t, err)

	cancel()

	value, err = db.Get(pair1.Key)
	assert.NoError(t, err)
	assert.Equal(t, value.Data, pair1.Value)
}

func TestDB_Watch(t *testing.T) {
	reset()

	watcher, err := db.Watch(pair1.Key)
	assert.NoError(t, err)

	defer watcher.Close()
	for i := 0; i < 10; i++ {
		assert.NoError(t, db.Set(pair1.Key, pair1.Value))

		value := <-watcher.Notify()
		assert.NotNil(t, value)
		assert.Equal(t, pair1.Value, *value.Value)
	}
}

func TestDB_WatchStrictMode(t *testing.T) {
	reset()

	nuts.EnableStrictMode()
	defer nuts.DisableStrictMode()

	watcher, err := db.Watch(pair1.Key)
	assert.NoError(t, err)
	defer watcher.Close()

	_, err = db.Watch(pair2.Key)
	// in strict mode, it will check whether the target Key exists.
	assert.Error(t, err)
}

func TestDB_WatchPrefix(t *testing.T) {
	reset()

	watcher := db.WatchPrefix([]byte("K"))
	defer func() {
		assert.NoError(t, watcher.Close())
	}()

	for _, node := range pairMatrix {
		assert.NoError(t, db.Set(node.Key, node.Value))

		value := <-watcher.Notify()
		assert.NotNil(t, value)
		assert.Equal(t, node.Value, *value.Value)
	}
}

func TestDB_Current(t *testing.T) {
	reset()

	assert.Equal(t, "", db.Current())

	newDB, err := db.Swap("bucket1")
	assert.NoError(t, err)
	assert.Equal(t, "bucket1", newDB.Current())
}

func TestDB_Snapshot(t *testing.T) {
	reset()

	for _, node := range pairMatrix {
		assert.NoError(t, db.Set(node.Key, node.Value))
	}

	reader, err := db.Snapshot()
	assert.NoError(t, err)
	assert.NotNil(t, reader)

	time.Sleep(100 * time.Millisecond) // wait db state change

	assert.NoError(t, db.Restore(reader))

	for _, node := range pairMatrix {
		value, err := db.Get(node.Key)
		assert.NoError(t, err)
		assert.Equal(t, node.Value, value.Data)
	}
}

func TestDB_Close(t *testing.T) {
	reset()

	assert.NoError(t, db.Close())

	// expect: timeout...
	ShouldTimeout(t, 1*time.Second, func() {
		db.Set(pair1.Key, pair1.Value)
	})
}

func TestDB_PrefixScan(t *testing.T) {
	reset()

	for _, node := range pairMatrix {
		assert.NoError(t, db.Set(node.Key, node.Value))
	}

	values, err := db.PrefixScan([]byte("K"), 0, 10)
	assert.NoError(t, err)
	assert.Len(t, values, len(pairMatrix))

	for i, value := range values {
		assert.Equal(t, pairMatrix[i].Value, value.Data)
	}
}

func TestDB_PrefixSearchScan(t *testing.T) {
	reset()

	for _, node := range pairMatrix {
		assert.NoError(t, db.Set(node.Key, node.Value))
	}

	values, err := db.PrefixSearchScan([]byte("K"), "", 0, 10)
	assert.NoError(t, err)
	assert.Len(t, values, len(pairMatrix))

	for i, value := range values {
		assert.Equal(t, pairMatrix[i].Value, value.Data)
	}
}

func TestDB_Set(t *testing.T) {
	reset()

	for _, node := range pairMatrix {
		assert.NoError(t, db.Set(node.Key, node.Value))
	}
}

func TestDB_SetWithTTL(t *testing.T) {
	reset()

	assert.NoError(t, db.SetWithTTL(pair1.Key, pair1.Value, 3))

	time.Sleep(2 * time.Second)
	value, err := db.Get(pair1.Key)
	assert.NoError(t, err)
	assert.NotNil(t, value.Data)
	assert.NotZero(t, value.TTL)

	time.Sleep(2 * time.Second) // waiting expired
	value, err = db.Get(pair1.Key)
	assert.Error(t, err)
	assert.Nil(t, value.Data)
	assert.Zero(t, value.TTL)
}

func TestDB_TrySetWithTTL(t *testing.T) {
	reset()

	assert.Error(t, db.TrySetWithTTL(pair1.Key, pair1.Value, 3)) // expect: error
	assert.NoError(t, db.TrySetWithTTL(pair2.Key, pair2.Value, 3))
	assert.Error(t, db.TrySetWithTTL(pair2.Key, pair2.Value, 3)) // expect: error

	time.Sleep(2 * time.Second)
	value, err := db.Get(pair2.Key)
	assert.NoError(t, err)
	assert.NotNil(t, value.Data)
	assert.NotZero(t, value.TTL)

	time.Sleep(2 * time.Second) // waiting expired
	value, err = db.Get(pair2.Key)
	assert.Error(t, err)
	assert.Nil(t, value.Data)
	assert.Zero(t, value.TTL)
}

func TestDB_TrySet(t *testing.T) {
	reset()

	assert.Error(t, db.TrySet(pair1.Key, pair1.Value)) // expect: error
	assert.NoError(t, db.TrySet(pair2.Key, pair2.Value))
	assert.Error(t, db.TrySet(pair2.Key, pair2.Value)) // expect: error
}

func TestDB_Del(t *testing.T) {
	reset()

	assert.NoError(t, db.Del(pair1.Key))
	assert.Error(t, db.Del(pair2.Key))

	assert.NoError(t, db.Set(pair2.Key, pair2.Value))
	assert.NoError(t, db.Del(pair2.Key))
}

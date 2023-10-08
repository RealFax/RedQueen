package nuts_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/RealFax/RedQueen/store/nuts"
)

var (
	keys = [][]byte{
		[]byte("KEY_1"),
		[]byte("KEY_2"),
		[]byte("KEY_3"),
		[]byte("KEY_4"),
		[]byte("KEY_5"),
		[]byte("KEY_6"),
		[]byte("KEY_7"),
		[]byte("KEY_8"),
		[]byte("KEY_9"),
		[]byte("KEY_10"),
	}
)

func TestWatcher(t *testing.T) {
	watcher := &nuts.Watcher{}

	child := watcher.Namespace("RedQueen")

	ctx, cancel := context.WithCancel(context.Background())

	wg := sync.WaitGroup{}
	for client := 0; client < 3; client++ {
		wg.Add(1)
		clientID := client
		go func() {
			t.Logf("Start recv, ClientID: %d", clientID)
			notify := child.Watch(keys[0])
			prefixNotify := child.WatchPrefix([]byte("KEY"))
			wg.Done()
			for {
				select {
				case val := <-notify.Values:
					t.Logf("ClientID: %d, Seq: %d, Timestamp: %d, TTL: %d, Key: %s, Value: %s",
						clientID,
						val.Seq,
						val.Timestamp,
						val.TTL,
						val.Key,
						*val.Value,
					)
				case val := <-prefixNotify.Values:
					t.Logf("[Prefix] ClientID: %d, Seq: %d, Timestamp: %d, TTL: %d, Key: %s, Value: %s",
						clientID,
						val.Seq,
						val.Timestamp,
						val.TTL,
						val.Key,
						*val.Value,
					)
				case <-ctx.Done():
					notify.Close()
					t.Logf("Stop recv, ClientID: %d", clientID)
					return
				}
			}
		}()
	}

	wg.Wait()
	for i := 0; i < 10; i++ {
		child.Update(keys[i], []byte("VALUE_"+strconv.Itoa(i)), 60)
		time.Sleep(time.Millisecond * 300)
	}

	cancel()

	time.Sleep(time.Second * 1)
}

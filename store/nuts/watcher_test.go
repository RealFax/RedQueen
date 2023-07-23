package nuts_test

import (
	"context"
	"encoding/json"
	"github.com/RealFax/RedQueen/store/nuts"
	"strconv"
	"testing"
	"time"
)

var (
	keys = [][]byte{
		[]byte("KEY_1"),
		[]byte("KEY_2"),
		[]byte("KEY_3"),
		[]byte("KEY_4"),
		[]byte("KEY_5"),
	}
)

func TestWatcher(t *testing.T) {
	watcher := &nuts.Watcher{}

	child := watcher.Namespace("RedQueen")

	ctx, cancel := context.WithCancel(context.Background())

	for client := 0; client < 3; client++ {
		clientID := client
		go func() {
			t.Logf("Start recv, ClientID: %d", clientID)
			notify := child.Watch(keys[0])
			for {
				select {
				case val := <-notify.Values:
					t.Logf("ClientID: %d, Seq: %d, Timestamp: %d, Data: %s", clientID, val.Seq, val.Timestamp, *val.Data)
				case <-ctx.Done():
					notify.Close()
					t.Logf("Stop recv, ClientID: %d", clientID)
					return
				}
			}
		}()
	}

	for i := 0; i < 10; i++ {
		child.Update(keys[0], []byte("VALUE_"+strconv.Itoa(i)))
		time.Sleep(time.Millisecond * 300)
	}

	out, _ := json.MarshalIndent(watcher.Metadata(), "", "\t")
	t.Logf("\n%s", out)
	cancel()

	time.Sleep(time.Second * 1)
}

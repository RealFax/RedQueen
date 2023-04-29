package nuts

import (
	"github.com/RealFax/RedQueen/store"
	"github.com/google/uuid"
	"sync"
	"time"
)

const WatcherNotifyValueSize = 1024

type WatcherNotify struct {
	unwatchedFunc func()
	Values        chan *store.WatchValue
	UUID          string
}

func (n *WatcherNotify) Close() error {
	n.unwatchedFunc()
	close(n.Values)
	return nil
}

type WatcherChannel struct {
	mu         sync.Mutex
	Seq        uint64
	LastUpdate int64
	Value      *[]byte
	Notify     sync.Map // map[string]*WatcherNotify
}

func (c *WatcherChannel) AddNotify(dest *WatcherNotify) {
	dest.unwatchedFunc = func() {
		c.Notify.Range(func(key, _ any) bool {
			if key.(string) == dest.UUID {
				c.Notify.Delete(key)
				return false
			}
			return true
		})
	}
	c.Notify.Store(dest.UUID, dest)
}

func (c *WatcherChannel) UpdateValue(val *[]byte) {
	var (
		seq       uint64
		timestamp int64
	)

	c.mu.Lock()
	c.Seq++
	c.LastUpdate = time.Now().UnixMilli()
	c.Value = val

	seq = c.Seq
	timestamp = c.LastUpdate
	c.mu.Unlock()

	c.Notify.Range(func(_, value any) bool {
		dest := value.(*WatcherNotify)
		// chan is full
		if len(dest.Values) == cap(dest.Values) {
			return true
		}
		dest.Values <- &store.WatchValue{
			Seq:       seq,
			Timestamp: timestamp,
			Data:      val,
		}
		return true
	})
}

type WatcherChild struct {
	Namespace string
	Channels  sync.Map // map[uint64]*WatcherChannel
}

func (c *WatcherChild) Watch(key []byte) *WatcherNotify {
	watchKey := WatchKey(key)

	channel, ok := c.Channels.Load(watchKey)
	if !ok {
		// init watcher channel
		channel = &WatcherChannel{}
		c.Channels.Store(watchKey, channel)
	}

	// create watcher notify
	notify := &WatcherNotify{
		Values: make(chan *store.WatchValue, WatcherNotifyValueSize),
		UUID:   uuid.New().String(),
	}

	// setup watcher notify
	channel.(*WatcherChannel).AddNotify(notify)

	return notify
}

func (c *WatcherChild) Update(key, value []byte) {
	watchKey := WatchKey(key)

	c.Channels.Range(func(key, channel any) bool {
		if watchKey == key {
			channel.(*WatcherChannel).UpdateValue(&value)
			return false
		}
		return true
	})
}

type Watcher struct {
	Namespaces sync.Map // map[string]*WatcherChild
}

func (w *Watcher) Namespace(namespace string) *WatcherChild {
	child, ok := w.Namespaces.Load(namespace)
	if !ok {
		child = &WatcherChild{Namespace: namespace}
		w.Namespaces.Store(namespace, child)
	}
	return child.(*WatcherChild)
}

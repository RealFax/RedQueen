package nuts

import (
	"encoding/json"
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/utils"
	"github.com/google/uuid"
	"sync"
	"time"
)

const WatcherNotifyValueSize = 1024

type WatcherMetadata map[string]map[uint64]map[string]int

func (m *WatcherMetadata) String() string {
	out, _ := json.MarshalIndent(m, "", "\t")
	return utils.Bytes2String(out)
}

type WatcherNotify struct {
	unwatchedFunc func()
	Values        chan *store.WatchValue
	UUID          string
}

func (n *WatcherNotify) Notify() chan *store.WatchValue {
	return n.Values
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

func (c *WatcherChannel) Get() map[string]int {
	m := make(map[string]int)
	c.Notify.Range(func(key, value any) bool {
		m[key.(string)] = len(value.(*WatcherNotify).Values)
		return true
	})
	return m
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

func (c *WatcherChild) Get() map[uint64]map[string]int {
	m := make(map[uint64]map[string]int)
	c.Channels.Range(func(key, value any) bool {
		m[key.(uint64)] = value.(*WatcherChannel).Get()
		return true
	})
	return m
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

	var valuePtr *[]byte
	if value != nil {
		valuePtr = &value
	}

	c.Channels.Range(func(key, channel any) bool {
		if watchKey == key {
			channel.(*WatcherChannel).UpdateValue(valuePtr)
			return false
		}
		return true
	})
}

type Watcher struct {
	Namespaces sync.Map // map[string]*WatcherChild
}

func (w *Watcher) Get() *WatcherMetadata {
	m := make(map[string]map[uint64]map[string]int)
	w.Namespaces.Range(func(key, value any) bool {
		m[key.(string)] = value.(*WatcherChild).Get()
		return true
	})
	return (*WatcherMetadata)(&m)
}

func (w *Watcher) Namespace(namespace string) *WatcherChild {
	child, ok := w.Namespaces.Load(namespace)
	if !ok {
		child = &WatcherChild{Namespace: namespace}
		w.Namespaces.Store(namespace, child)
	}
	return child.(*WatcherChild)
}

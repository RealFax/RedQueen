package nuts

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/RealFax/RedQueen/common/hack"
	"github.com/RealFax/RedQueen/store"
)

const WatcherNotifyValueSize = 1024

type WatcherMetadata map[string]map[uint64]map[string]int

func (m *WatcherMetadata) String() string {
	out, _ := json.MarshalIndent(m, "", "\t")
	return hack.Bytes2String(out)
}

type WatcherNotify struct {
	state         atomic.Bool
	unwatchedFunc func()
	Values        chan *store.WatchValue
	UUID          string
}

func (n *WatcherNotify) Notify() chan *store.WatchValue {
	return n.Values
}

func (n *WatcherNotify) Close() error {
	if !n.state.Load() {
		return errors.New("watcher notify has closed")
	}
	n.state.Store(false)
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

func (c *WatcherChannel) Metadata() map[string]int {
	m := make(map[string]int)
	c.Notify.Range(func(key, value any) bool {
		m[key.(string)] = len(value.(*WatcherNotify).Values)
		return true
	})
	return m
}

func (c *WatcherChannel) AddNotify(dest *WatcherNotify) {
	dest.state = atomic.Bool{}
	dest.state.Store(true)
	dest.unwatchedFunc = func() {
		c.Notify.Delete(dest.UUID)
	}
	c.Notify.Store(dest.UUID, dest)
}

func (c *WatcherChannel) UpdateValue(val *[]byte, ttl uint32) {
	c.mu.Lock()
	c.Seq++
	c.LastUpdate = time.Now().UnixMilli()
	c.Value = val

	seq := c.Seq
	timestamp := c.LastUpdate
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
			TTL:       ttl,
			Data:      val,
		}
		return true
	})
}

type WatcherChild struct {
	Namespace string
	Channels  sync.Map // map[uint64]*WatcherChannel
}

func (c *WatcherChild) Metadata() map[uint64]map[string]int {
	m := make(map[uint64]map[string]int)
	c.Channels.Range(func(key, value any) bool {
		m[key.(uint64)] = value.(*WatcherChannel).Metadata()
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

func (c *WatcherChild) Update(key, value []byte, ttl uint32) {
	channel, ok := c.Channels.Load(WatchKey(key))
	if !ok {
		return
	}

	var valuePtr *[]byte
	if value != nil {
		valuePtr = &value
	}

	channel.(*WatcherChannel).UpdateValue(valuePtr, ttl)
}

type Watcher struct {
	Namespaces sync.Map // map[string]*WatcherChild
}

func (w *Watcher) Metadata() *WatcherMetadata {
	m := make(map[string]map[uint64]map[string]int)
	w.Namespaces.Range(func(key, value any) bool {
		m[key.(string)] = value.(*WatcherChild).Metadata()
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

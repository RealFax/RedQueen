package nuts

import (
	"bytes"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/RealFax/RedQueen/store"
)

const WatcherNotifyValueSize = 1024

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
	mu          sync.Mutex
	PrefixWatch bool
	Seq         uint64
	LastUpdate  int64
	Value       *[]byte
	Prefix      []byte
	Notify      sync.Map // map[string]*WatcherNotify
}

func (c *WatcherChannel) AddNotify(dest *WatcherNotify) {
	dest.state = atomic.Bool{}
	dest.state.Store(true)
	dest.unwatchedFunc = func() {
		c.Notify.Delete(dest.UUID)
	}
	c.Notify.Store(dest.UUID, dest)
}

func (c *WatcherChannel) UpdateValue(key []byte, value *[]byte, ttl uint32) {
	if c.PrefixWatch && !bytes.HasPrefix(key, c.Prefix) {
		return
	}

	c.mu.Lock()
	c.Seq++
	c.LastUpdate = time.Now().UnixMilli()
	c.Value = value

	seq := c.Seq
	timestamp := c.LastUpdate
	c.mu.Unlock()

	watchValue := &store.WatchValue{
		Seq:       seq,
		Timestamp: timestamp,
		TTL:       ttl,
		Key:       key,
		Value:     value,
	}

	c.Notify.Range(func(_, value any) bool {
		dest := value.(*WatcherNotify)
		// chan is full
		if len(dest.Values) == cap(dest.Values) {
			return true
		}
		dest.Values <- watchValue
		return true
	})
}

type WatcherChild struct {
	Namespace      string
	Channels       sync.Map // map[uint64]*WatcherChannel
	PrefixChannels sync.Map // map[string]*WatcherChannel
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

func (c *WatcherChild) WatchPrefix(prefix []byte) *WatcherNotify {
	prefixKey := PrefixKey(prefix)

	channel, ok := c.PrefixChannels.Load(prefixKey)
	if !ok {
		// init prefix watcher channel
		channel = &WatcherChannel{PrefixWatch: true, Prefix: prefix}
		c.PrefixChannels.Store(prefixKey, channel)
	}

	// create prefix watcher notify
	notify := &WatcherNotify{
		Values: make(chan *store.WatchValue, WatcherNotifyValueSize),
		UUID:   uuid.New().String(),
	}

	// setup prefix watcher notify
	channel.(*WatcherChannel).AddNotify(notify)

	return notify
}

func (c *WatcherChild) Update(key, value []byte, ttl uint32) {
	var valuePtr *[]byte
	if value != nil {
		valuePtr = &value
	}

	// update prefix channels value
	c.PrefixChannels.Range(func(_, value any) bool {
		channel, _ := value.(*WatcherChannel)
		channel.UpdateValue(key, valuePtr, ttl)
		return true
	})

	channel, ok := c.Channels.Load(WatchKey(key))
	if !ok {
		return
	}

	channel.(*WatcherChannel).UpdateValue(key, valuePtr, ttl)
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

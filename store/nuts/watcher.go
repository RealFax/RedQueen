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

const WatcherNotifyBufferSize = 32

type WatcherNotifier struct {
	state         atomic.Bool
	unwatchedFunc func()
	Values        chan *store.WatchValue
	UUID          string
}

func (n *WatcherNotifier) Notify() chan *store.WatchValue {
	return n.Values
}

func (n *WatcherNotifier) Close() error {
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
	sequence    uint64
	lastUpdate  int64
	Value       *[]byte
	Prefix      []byte
	Notify      sync.Map // map[string]*WatcherNotifier
}

func (c *WatcherChannel) AddNotifier(dest *WatcherNotifier) {
	dest.state = atomic.Bool{}
	dest.state.Store(true)
	dest.unwatchedFunc = func() {
		c.Notify.Delete(dest.UUID)
	}
	c.Notify.Store(dest.UUID, dest)
}

func (c *WatcherChannel) TryUpdate(key []byte, value *[]byte, ttl uint32) {
	if c.PrefixWatch && !bytes.HasPrefix(key, c.Prefix) {
		return
	}

	c.mu.Lock()
	c.sequence++
	c.lastUpdate = time.Now().UnixMilli()
	c.Value = value

	seq := c.sequence
	timestamp := c.lastUpdate
	c.mu.Unlock()

	watchValue := &store.WatchValue{
		Seq:       seq,
		Timestamp: timestamp,
		TTL:       ttl,
		Key:       key,
		Value:     value,
	}

	c.Notify.Range(func(_, value any) bool {
		dest := value.(*WatcherNotifier)
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

func (c *WatcherChild) Watch(key []byte) *WatcherNotifier {
	watchKey := WatchKey(key)

	channel, ok := c.Channels.Load(watchKey)
	if !ok {
		// init watcher channel
		_channel := &WatcherChannel{}
		if c.Channels.CompareAndSwap(watchKey, nil, _channel) {
			channel = _channel
		}
	}

	// create watcher notify
	notify := &WatcherNotifier{
		Values: make(chan *store.WatchValue, WatcherNotifyBufferSize),
		UUID:   uuid.New().String(),
	}

	// setup watcher notify
	channel.(*WatcherChannel).AddNotifier(notify)

	return notify
}

func (c *WatcherChild) WatchPrefix(prefix []byte) *WatcherNotifier {
	prefixKey := PrefixKey(prefix)

	channel, ok := c.PrefixChannels.Load(prefixKey)
	if !ok {
		// init prefix watcher channel
		_channel := &WatcherChannel{PrefixWatch: true, Prefix: prefix}
		if c.PrefixChannels.CompareAndSwap(prefixKey, nil, _channel) {
			channel = _channel
		}
	}

	// create prefix watcher notify
	notify := &WatcherNotifier{
		Values: make(chan *store.WatchValue, WatcherNotifyBufferSize),
		UUID:   uuid.New().String(),
	}

	// setup prefix watcher notify
	channel.(*WatcherChannel).AddNotifier(notify)

	return notify
}

func (c *WatcherChild) Update(key, value []byte, ttl uint32) {
	var pValue *[]byte
	if value != nil {
		pValue = &value
	}

	// update prefix channels value
	c.PrefixChannels.Range(func(_, value any) bool {
		channel, _ := value.(*WatcherChannel)
		channel.TryUpdate(key, pValue, ttl)
		return true
	})

	channel, ok := c.Channels.Load(WatchKey(key))
	if !ok {
		return
	}

	channel.(*WatcherChannel).TryUpdate(key, pValue, ttl)
}

type Watcher struct {
	Namespaces sync.Map // map[string]*WatcherChild
}

func (w *Watcher) UseTarget(namespace string) *WatcherChild {
	child, ok := w.Namespaces.Load(namespace)
	if !ok {
		_child := &WatcherChild{Namespace: namespace}
		if w.Namespaces.CompareAndSwap(namespace, nil, child) {
			child = _child
		}
	}
	return child.(*WatcherChild)
}

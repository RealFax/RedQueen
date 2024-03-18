package balancer

import (
	"github.com/RealFax/RedQueen/pkg/expr"
	"github.com/pkg/errors"
	"sync/atomic"
)

type roundRobinBalance[K comparable, V any] struct {
	current atomic.Int32
	*store[K, V]
}

func (b *roundRobinBalance[K, V]) Next() (V, error) {
	size := b.Size()
	if size == 0 {
		return expr.Zero[V](), errors.New("empty load balance list")
	}
	next := b.current.Add(1) % size
	b.current.CompareAndSwap(b.current.Load(), next)

	b.mu.RLock()
	nextValue := b.nodes[next].Value()
	b.mu.RUnlock()
	return nextValue, nil
}

func (b *roundRobinBalance[K, V]) Remove(key K) bool {
	if !b.store.Remove(key) {
		return false
	}
	b.current.Add(-1)
	return true
}

func NewRoundRobin[K comparable, V any]() Balancer[K, V] {
	return &roundRobinBalance[K, V]{store: newLoadBalanceStore[K, V]()}
}

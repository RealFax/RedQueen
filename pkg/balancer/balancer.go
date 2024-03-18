package balancer

import (
	"github.com/RealFax/RedQueen/pkg/maputil"
	"slices"
	"sync"
)

type Pod[K comparable, V any] interface {
	Key() K
	Value() V
}

type Balancer[K comparable, V any] interface {
	Size() int32
	Append(node ...Pod[K, V])
	Remove(key K) bool
	Next() (V, error)
}

type store[K comparable, V any] struct {
	mu sync.RWMutex

	size   int32
	filter *maputil.Map[K, struct{}]
	nodes  []Pod[K, V]
}

func (s *store[K, V]) Size() int32 {
	s.mu.RLock()
	size := s.size
	s.mu.RUnlock()
	return size
}

func (s *store[K, V]) Append(nodes ...Pod[K, V]) {
	for i, node := range nodes {
		if _, ok := s.filter.Load(node.Key()); ok {
			nodes = append(nodes[:i], nodes[i+1:]...)
			continue
		}
		s.filter.Store(node.Key(), struct{}{})
	}

	s.mu.Lock()
	s.nodes = append(s.nodes, nodes...)
	s.size += int32(len(nodes))
	s.mu.Unlock()
}

func (s *store[K, V]) Remove(key K) bool {
	if _, ok := s.filter.LoadAndDelete(key); !ok {
		return false
	}

	s.mu.Lock()
	s.nodes = slices.DeleteFunc(s.nodes, func(n Pod[K, V]) bool {
		return n.Key() == key
	})
	s.size -= 1
	s.mu.Unlock()

	return true
}

func newLoadBalanceStore[K comparable, V any]() *store[K, V] {
	return &store[K, V]{
		filter: maputil.New[K, struct{}](),
		nodes:  make([]Pod[K, V], 0, 64),
	}
}

package balancer

import (
	"github.com/RealFax/RedQueen/pkg/maputil"
	"slices"
	"sync"
)

type Node[K comparable, V any] interface {
	Key() K
	Value() V
}

type LoadBalance[K comparable, V any] interface {
	Size() int32
	Append(node ...Node[K, V])
	Remove(key K) bool
	Next() (V, error)
}

type loadBalanceStore[K comparable, V any] struct {
	rwm sync.RWMutex

	size   int32
	filter *maputil.Map[K, struct{}]
	nodes  []Node[K, V]
}

func (s *loadBalanceStore[K, V]) Size() int32 {
	s.rwm.RLock()
	size := s.size
	s.rwm.RUnlock()
	return size
}

func (s *loadBalanceStore[K, V]) Append(nodes ...Node[K, V]) {
	for i, node := range nodes {
		if _, ok := s.filter.Load(node.Key()); ok {
			nodes = append(nodes[:i], nodes[i+1:]...)
			continue
		}
		s.filter.Store(node.Key(), struct{}{})
	}

	s.rwm.Lock()
	s.nodes = append(s.nodes, nodes...)
	s.size += int32(len(nodes))
	s.rwm.Unlock()
}

func (s *loadBalanceStore[K, V]) Remove(key K) bool {
	if _, ok := s.filter.LoadAndDelete(key); !ok {
		return false
	}

	s.rwm.Lock()
	s.nodes = slices.DeleteFunc(s.nodes, func(n Node[K, V]) bool {
		return n.Key() == key
	})
	s.size -= 1
	s.rwm.Unlock()

	return true
}

func newLoadBalanceStore[K comparable, V any]() *loadBalanceStore[K, V] {
	return &loadBalanceStore[K, V]{
		filter: maputil.New[K, struct{}](),
		nodes:  make([]Node[K, V], 0, 64),
	}
}

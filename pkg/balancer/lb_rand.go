package balancer

import (
	"crypto/rand"
	"github.com/RealFax/RedQueen/pkg/expr"
	"github.com/pkg/errors"
	"math/big"
)

type randomBalance[K comparable, V any] struct {
	*store[K, V]
}

func (b *randomBalance[K, V]) Next() (V, error) {
	size := b.Size()
	if size == 0 {
		return expr.Zero[V](), errors.New("empty load balance list")
	}
	idx, _ := rand.Int(rand.Reader, big.NewInt(int64(size)))

	b.mu.RLock()
	nextValue := b.nodes[idx.Int64()].Value()
	b.mu.RUnlock()
	return nextValue, nil
}

func NewRandom[K comparable, V any]() Balancer[K, V] {
	return &randomBalance[K, V]{store: newLoadBalanceStore[K, V]()}
}

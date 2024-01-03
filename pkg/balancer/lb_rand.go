package balancer

import (
	"crypto/rand"
	"github.com/pkg/errors"
	"math/big"
)

type randomBalance[K comparable, V any] struct {
	*loadBalanceStore[K, V]
}

func (b *randomBalance[K, V]) Next() (V, error) {
	size := b.Size()
	if size == 0 {
		var empty V
		return empty, errors.New("empty load balance list")
	}
	idx, _ := rand.Int(rand.Reader, big.NewInt(int64(size)))

	b.rwm.RLock()
	nextValue := b.nodes[idx.Int64()].Value()
	b.rwm.RUnlock()
	return nextValue, nil
}

func NewRandom[K comparable, V any]() LoadBalance[K, V] {
	return &randomBalance[K, V]{newLoadBalanceStore[K, V]()}
}

package syncx

import (
	"sync"
	"sync/atomic"
)

type PoolVarWrap[T any] struct {
	freed atomic.Bool
	pool  *Pool[T]
	val   T
}

func (pv *PoolVarWrap[T]) Free() {
	if pv.freed.Load() {
		panic("value has freed")
	}
	pv.pool.Free(pv.val)
	pv.freed.Swap(true)
}

func (pv *PoolVarWrap[T]) Val() T {
	if pv.freed.Load() {
		panic("value has freed")
	}
	return pv.val
}

type Pool[T any] struct {
	pool    sync.Pool
	onAlloc func(val T)
	onFree  func(val T)
}

func (p *Pool[T]) Alloc() *PoolVarWrap[T] {
	val := p.pool.Get().(T)
	if p.onAlloc != nil {
		p.onAlloc(val)
	}

	return &PoolVarWrap[T]{
		pool: p,
		val:  val,
	}
}

func (p *Pool[T]) Free(val T) {
	if p.onFree != nil {
		p.onFree(val)
	}
	p.pool.Put(val)
}

// NewPool
//
// deprecated: syncx has been deprecated due to performance issues
func NewPool[T any](new func() T, onAlloc func(val T), onFree func(val T)) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return new()
			},
		},
		onAlloc: onAlloc,
		onFree:  onFree,
	}
}

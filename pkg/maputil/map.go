// map util is a wrapper of sync.Map, type support
// doc see: https://pkg.go.dev/sync#Map

package maputil

import (
	"github.com/RealFax/RedQueen/pkg/hack"
	json "github.com/json-iterator/go"
	"sync"
)

type Map[K comparable, V any] struct {
	m sync.Map
}

func (v *Map[K, V]) assert(value any, exist bool) (V, bool) {
	if !exist {
		var empty V
		return empty, false
	}

	assertVal, ok := value.(V)
	return assertVal, ok
}

// ---- sync.Map wrapper ----

func (v *Map[K, V]) Load(key K) (V, bool) {
	return v.assert(v.m.Load(key))
}

func (v *Map[K, V]) Store(key K, value V) {
	v.m.Store(key, value)
}

func (v *Map[K, V]) LoadOrStore(key K, value V) (V, bool) {
	return v.assert(v.m.LoadOrStore(key, value))
}

func (v *Map[K, V]) LoadAndDelete(key K) (V, bool) {
	return v.assert(v.m.LoadAndDelete(key))
}

func (v *Map[K, V]) Delete(key K) {
	v.m.Delete(key)
}

func (v *Map[K, V]) Swap(key K, value V) (V, bool) {
	return v.assert(v.m.Swap(key, value))
}

func (v *Map[K, V]) CompareAndSwap(key K, old, new V) bool {
	return v.m.CompareAndSwap(key, old, new)
}

func (v *Map[K, V]) CompareAndDelete(key K, old V) bool {
	return v.m.CompareAndDelete(key, old)
}

func (v *Map[K, V]) Range(f func(key K, value V) bool) {
	v.m.Range(func(key, value any) bool {
		return f(key.(K), value.(V))
	})
}

// ---- advanced func ----

func (v *Map[K, V]) Exist(key K) bool {
	_, exist := v.m.Load(key)
	return exist
}

func (v *Map[K, V]) Map() map[K]V {
	m := make(map[K]V)
	v.Range(func(key K, value V) bool {
		m[key] = value
		return true
	})
	return m
}

func (v *Map[K, V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.Map())
}

func (v *Map[K, V]) UnmarshalJSON(b []byte) error {
	s := hack.Bytes2String(b)
	if s == "null" || s == "" {
		return nil
	}

	var m map[K]V
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}

	*v = *Clone(m)
	return nil
}

func (v *Map[K, V]) Copy(dst *Map[K, V]) {
	v.Range(func(key K, value V) bool {
		dst.Store(key, value)
		return true
	})
}

func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{}
}

// Clone returns a copy of src.
// this is a shallow clone the new keys and values are set using ordinary assignment.
func Clone[M ~map[K]V, K comparable, V any](src M) *Map[K, V] {
	dst := New[K, V]()
	for k, v := range src {
		dst.Store(k, v)
	}
	return dst
}

// Copy copies all key/value pairs in src adding them to dst.
// When a key in src is already present in dst,
// the value in dst will be overwritten by the value associated
// with the key in src.
func Copy[M ~map[K]V, K comparable, V any](src *Map[K, V], dst M) {
	src.Range(func(key K, value V) bool {
		dst[key] = value
		return true
	})
}

// Keys copy all keys in the map as a slice
func Keys[M ~map[K]V, K comparable, V any](src M) []K {
	var (
		off  int
		size = len(src)
		keys = make([]K, size)
	)
	for key := range src {
		keys[off] = key
		off++
		if off >= size {
			break
		}
	}
	return keys[:off]
}

// KeysFunc copy all keys where fc returns true from src to the slice
func KeysFunc[M ~map[K]V, K comparable, V any](src M, fc func(K, V) bool) []K {
	var (
		off  int
		size = len(src)
		keys = make([]K, 0, size)
	)
	for key, value := range src {
		if fc(key, value) {
			keys = append(keys, key)
		}
		off++
		if off >= size {
			break
		}
	}
	return keys
}

// AssignSlice convert a slice to a map
func AssignSlice[S ~[]K, K comparable](src S) map[K]struct{} {
	m := make(map[K]struct{})
	for i := range src {
		m[src[i]] = struct{}{}
	}
	return m
}

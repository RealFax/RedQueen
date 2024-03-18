package balancer

type weightedBalance[K comparable, V any] struct {
	*store[K, V]
}

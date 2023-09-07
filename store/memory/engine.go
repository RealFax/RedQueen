package memory

import (
	"sync"
	"sync/atomic"
)

type (
	PodState uint32

	PodMetadata struct {
		// pod state
		state         PodState
		NamespaceSize *uint32
		KeySize       uint32
		ValueSize     uint32
		TTL           uint32
		// LastUpdated is a timestamp
		LastUpdated uint64
		Namespace   *string
	}

	Pod struct {
		Meta *PodMetadata
		Key  []byte
		Data []byte
	}

	NodeMetadata struct {
		NamespaceSize *uint32
		PodSize       uint64
		Namespace     *string
	}

	Node struct {
		rwm      sync.RWMutex
		Meta     *NodeMetadata
		Mappings map[uint64]*Pod
	}
)

const (
	PodStateOk PodState = iota
)

// ---- PodMetadata ----

func (m *PodMetadata) State() PodState {
	stateAddr := (*uint32)(&m.state)
	return PodState(atomic.LoadUint32(stateAddr))
}

func (m *PodMetadata) EqualState(state PodState) bool {
	return m.State() == state
}

func (m *PodMetadata) PayloadSize() int64 {
	return int64(atomic.LoadUint32(m.NamespaceSize)) +
		int64(atomic.LoadUint32(&m.KeySize)) +
		int64(atomic.LoadUint32(&m.ValueSize))
}

// ---- Pod ----

func (p *Pod) MapKey() uint64 {
	return KeySum(p.Key)
}

// ---- NodeMetadata ----

func (m *NodeMetadata) SetNamespace(namespace string) {
	*m.Namespace = namespace
	atomic.StoreUint32(m.NamespaceSize, uint32(len(namespace)))
}

// ---- Node ----

func (n *Node) SetNamespace(namespace string) {
	n.rwm.Lock()
	n.Meta.SetNamespace(namespace)
	n.rwm.Unlock()
}

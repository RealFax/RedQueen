package rqd

import (
	"context"
	"github.com/RealFax/RedQueen/internal/rqd/store"
	"io"
	"net"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/pkg/errors"
)

type Raft struct {
	bootstrap bool
	term      uint64 // [ATOMIC]
	clusters  []raft.Server

	cfg           *raft.Config
	fsm           raft.FSM
	logStore      raft.LogStore
	stableStore   raft.StableStore
	snapshotStore raft.SnapshotStore
	transport     raft.Transport
	ctx           context.Context

	*raft.Raft
}

func (r *Raft) AddCluster(id raft.ServerID, addr raft.ServerAddress) error {
	return r.AddVoter(id, addr, 0, 30*time.Second).Error()
}

func (r *Raft) Term() uint64 {
	return atomic.LoadUint64((*uint64)(unsafe.Pointer(r.Raft)))
}

type RaftServerOption func(*Raft) error

func RaftWithEmpty() RaftServerOption { return func(r *Raft) error { return nil } }

func RaftWithContext(ctx context.Context) RaftServerOption {
	return func(r *Raft) error {
		r.ctx = ctx
		return nil
	}
}

func RaftWithBootstrap() RaftServerOption {
	return func(r *Raft) error {
		r.bootstrap = true
		return nil
	}
}

func RaftWithClusters(clusters []raft.Server) RaftServerOption {
	return func(r *Raft) error {
		r.clusters = clusters
		return nil
	}
}

func RaftWithConfig(cfg *raft.Config) RaftServerOption {
	return func(r *Raft) error {
		r.cfg = cfg
		return nil
	}
}

func RaftWithStdFSM(store store.Store) RaftServerOption {
	return func(r *Raft) error {
		r.fsm = &FSM{
			Term:     &r.term,
			Handlers: NewFSMHandlers(store),
			Store:    store,
		}
		return nil
	}
}

func RaftWithBoltLogStore(path string) RaftServerOption {
	return func(r *Raft) (err error) {
		r.logStore, err = raftboltdb.NewBoltStore(path)
		if err != nil {
			return errors.Wrap(err, "bolt-log-actions")
		}
		return
	}
}

func RaftWithStdStableStore(store store.Store) RaftServerOption {
	return func(r *Raft) (err error) {
		r.stableStore, err = NewStableStore(store)
		if err != nil {
			return errors.Wrap(err, "std-stable-actions")
		}
		return
	}
}

func RaftWithFileSnapshotStore(path string, retain int, logOut io.Writer) RaftServerOption {
	return func(r *Raft) (err error) {
		r.snapshotStore, err = raft.NewFileSnapshotStore(path, retain, logOut)
		if err != nil {
			return errors.Wrap(err, "file-snapshot-actions")
		}
		return
	}
}

func RaftWithTCPTransport(addr string, maxPool int, timeout time.Duration, logOut io.Writer) RaftServerOption {
	return func(r *Raft) error {
		tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return errors.Wrap(err, "resolve-tcp-addr")
		}
		if r.transport, err = raft.NewTCPTransport(addr, tcpAddr, maxPool, timeout, logOut); err != nil {
			return errors.Wrap(err, "tcp-transport")
		}
		return nil
	}
}

func NewRaftWithOptions(opts ...RaftServerOption) (*Raft, error) {
	var (
		err error
		r   = &Raft{}
	)

	for _, opt := range opts {
		if err = opt(r); err != nil {
			return nil, err
		}
	}

	if r.Raft, err = raft.NewRaft(r.cfg, r.fsm, r.logStore, r.stableStore, r.snapshotStore, r.transport); err != nil {
		return nil, err
	}

	if r.bootstrap {
		if fErr := r.BootstrapCluster(raft.Configuration{
			Servers: r.clusters,
		}); fErr.Error() != nil {
			return nil, fErr.Error()
		}
	}

	return r, nil
}

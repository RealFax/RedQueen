package RedQueen

import (
	"context"
	"github.com/RealFax/RedQueen/config"
	"github.com/RealFax/RedQueen/locker"
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/store/nuts"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
	"time"
)

type Server struct {
	term      uint64 // [ATOMIC]
	clusterID string

	store         store.Store
	lockerBackend locker.Backend

	raft *Raft
}

func (s *Server) currentNamespace(namespace *string) (store.Namespace, error) {
	var (
		err      error
		storeAPI store.Namespace = s.store
	)

	if namespace != nil {
		if storeAPI, err = s.store.Namespace(*namespace); err != nil {
			return nil, err
		}
	}

	return storeAPI, nil
}

func (s *Server) raftApply(ctx context.Context, timeout time.Duration, lp *LogPayload) (raft.ApplyFuture, error) {
	cmd, err := msgpack.Marshal(lp)
	if err != nil {
		return nil, err
	}

	var (
		ch = make(chan raft.ApplyFuture, 1)
	)

	go func() {
		ch <- s.raft.Apply(cmd, timeout)
		close(ch)
	}()

	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context canceled")
	case x := <-ch:
		err = x.Error()
		return x, err
	}
}

func newNutsStore(cfg config.Store) (store.Store, error) {
	if cfg.Nuts.StrictMode {
		nuts.EnableStrictMode()
	} else {
		nuts.DisableStrictMode()
	}

	return nuts.New(nuts.Config{
		NodeNum: cfg.Nuts.NodeNum,
		Sync:    cfg.Nuts.Sync,
		DataDir: cfg.Nuts.DataDir,
	})
}

func newStoreBackend(cfg config.Store) (store.Store, error) {
	handle, ok := map[config.EnumStoreBackend]func(config.Store) (store.Store, error){
		config.StoreBackendNuts: newNutsStore,
	}[cfg.Backend]
	if !ok {
		return nil, errors.New("unsupported store backend")
	}
	return handle(cfg)
}

func NewServer(cfg config.Config) (*Server, error) {

	var (
		server = Server{
			clusterID: cfg.Node.ID,
		}
		err error
	)

	if server.store, err = newStoreBackend(cfg.Store); err != nil {
		return nil, err
	}

	// init server raft
	if server.raft, err = NewRaft(RaftConfig{
		ServerID:              cfg.Node.ID,
		Addr:                  cfg.Node.ListenPeerAddr,
		BoltStorePath:         cfg.Node.DataDir + "/bolt",
		FileSnapshotStorePath: cfg.Node.DataDir,
		FSM:                   NewFSM(server.store),
		Clusters: func() []raft.Server {
			if !cfg.Env().FirstRun() {
				return nil
			}

			clusters := make([]raft.Server, len(cfg.Cluster.Bootstrap))
			for i, v := range cfg.Cluster.Bootstrap {
				clusters[i] = raft.Server{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(v.Name),
					Address:  raft.ServerAddress(v.PeerAddr),
				}
			}

			return clusters
		}(),
	}); err != nil {
		return nil, err
	}

	server.lockerBackend = NewLockerBackend(server.store, server.raftApply)

	return &server, nil
}

package RedQueen

import (
	"context"
	"github.com/RealFax/RedQueen/config"
	"github.com/RealFax/RedQueen/locker"
	"github.com/RealFax/RedQueen/store"
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

	cfg  *config.Config
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

func (s *Server) ListenClient() error {

	return nil
}

func (s *Server) Close() (err error) {
	if err = s.raft.Shutdown().Error(); err != nil {
		return
	}
	return
}

func NewServer(cfg *config.Config) (*Server, error) {

	var (
		server = Server{
			clusterID: cfg.Node.ID,
			cfg:       cfg,
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

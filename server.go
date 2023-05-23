package red

import (
	"context"
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/config"
	"github.com/RealFax/RedQueen/locker"
	"github.com/RealFax/RedQueen/store"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"net"
	"sync"
	"time"
)

type Server struct {
	term      uint64 // [ATOMIC]
	clusterID string

	cfg *config.Config

	store         store.Store
	lockerBackend locker.Backend

	raft       *Raft
	grpcServer *grpc.Server

	stateNotify sync.Map // map[string]chan bool

	serverpb.UnimplementedKVServer
	serverpb.UnimplementedLockerServer
	serverpb.UnimplementedRedQueenServer
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

func (s *Server) applyLog(ctx context.Context, p *serverpb.RaftLogPayload, timeout time.Duration) (raft.ApplyFuture, error) {
	cmd, err := proto.Marshal(p)
	if err != nil {
		return nil, errors.Wrap(err, "marshal raft log error")
	}

	ch := make(chan raft.ApplyFuture, 1)

	go func() {
		ch <- s.raft.Apply(cmd, timeout)
		close(ch)
	}()

	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context done")
	case v := <-ch:
		return v, v.Error()
	}
}

func (s *Server) _stateUpdater() {
	for {
		select {
		case state := <-s.raft.LeaderCh():
			s.stateNotify.Range(func(_, val any) bool {
				//ch, ok := val.(chan bool)
				//if !ok {
				//	return true
				//}
				//if len(ch) == cap(ch) {
				//	return true
				//}
				val.(chan bool) <- state
				return true
			})
		}
	}
}

func (s *Server) ListenServer() error {
	listener, err := net.Listen("tcp", s.cfg.Node.ListenClientAddr)
	if err != nil {
		return err
	}
	return s.grpcServer.Serve(listener)
}

func (s *Server) Close() (err error) {
	if err = s.raft.Shutdown().Error(); err != nil {
		return
	}
	s.grpcServer.Stop()
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

	// init server store backend
	if server.store, err = newStoreBackend(cfg.Store); err != nil {
		return nil, err
	}

	// init server grpc
	server.grpcServer = grpc.NewServer()
	serverpb.RegisterKVServer(server.grpcServer, &server)
	serverpb.RegisterLockerServer(server.grpcServer, &server)
	serverpb.RegisterRedQueenServer(server.grpcServer, &server)

	// init server raft
	if server.raft, err = NewRaft(cfg.Env().FirstRun(), RaftConfig{
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

	// init distributed lock backend
	server.lockerBackend = NewLockerBackend(server.store, server.raftApply)

	// start daemon service
	go server._stateUpdater()

	return &server, nil
}

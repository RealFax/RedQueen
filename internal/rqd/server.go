package rqd

import (
	"bytes"
	"context"
	"github.com/RealFax/RedQueen/internal/rqd/store"
	"github.com/RealFax/RedQueen/pkg/dlocker"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/config"
)

var (
	bufferPool = sync.Pool{New: func() any {
		return &bytes.Buffer{}
	}}
)

type Server struct {
	ctx    context.Context
	cancel context.CancelCauseFunc

	cfg *config.Config

	store         store.Store
	lockerBackend dlocker.Backend
	logApplyer    RaftApply

	raft        *Raft
	grpcServer  *grpc.Server
	pprofServer *pprofServer

	stateNotify sync.Map // map[string]chan bool

	clusterID string

	serverpb.UnimplementedKVServer
	serverpb.UnimplementedLockerServer
	serverpb.UnimplementedRedQueenServer
}

func (s *Server) trySwapContext(namespace *string) (store.Actions, error) {
	var (
		err     error
		actions store.Actions = s.store
	)

	if namespace != nil {
		if actions, err = s.store.Swap(*namespace); err != nil {
			return nil, err
		}
	}

	return actions, nil
}

func (s *Server) applyLog(ctx context.Context, p *serverpb.RaftLogPayload, timeout time.Duration) error {
	if err := s.logApplyer.Apply(&ctx, p, timeout); err != nil {
		if errors.Is(err, ErrApplyLogTimeTravelDone) || errors.Is(err, ErrApplyLogDone) {
			return nil
		}
		return err
	}
	// waiting response
	<-ctx.Done()
	if errors.Is(context.Cause(ctx), ErrApplyLogDone) {
		return nil
	}

	return ctx.Err()
}

func (s *Server) _stateUpdater() {
	for {
		select {
		case <-s.ctx.Done():
			return
		case state := <-s.raft.LeaderCh():
			s.stateNotify.Range(func(_, val any) bool {
				val.(chan bool) <- state
				return true
			})
		}
	}
}

func (s *Server) ListenAndServe() error {
	listener, err := net.Listen("tcp", s.cfg.Node.ListenClientAddr)
	if err != nil {
		return err
	}
	return s.grpcServer.Serve(listener)
}

func (s *Server) Close() (err error) {
	s.cancel(errors.New("server close"))

	if err = s.raft.Shutdown().Error(); err != nil {
		return
	}
	s.grpcServer.Stop()

	if s.cfg.PPROF {
		err = s.pprofServer.Close()
	}
	return
}

func NewServer(cfg *config.Config) (*Server, error) {
	var (
		err    error
		server = &Server{
			clusterID: cfg.Node.ID,
			cfg:       cfg,
		}
	)
	server.ctx, server.cancel = context.WithCancelCause(context.Background())

	if cfg.Misc.PPROF {
		server.pprofServer, err = newPprofServer()
		if err != nil {
			return nil, errors.Wrap(err, "pprof server")
		}
		go server.pprofServer.Run()
	}

	// init server actions backend
	if server.store, err = newStoreBackend(cfg.Store, cfg.Node.DataDir); err != nil {
		return nil, err
	}

	// init server grpc
	server.grpcServer = grpc.NewServer()
	serverpb.RegisterKVServer(server.grpcServer, server)
	serverpb.RegisterLockerServer(server.grpcServer, server)
	serverpb.RegisterRedQueenServer(server.grpcServer, server)

	// init server raft
	if server.raft, err = NewRaftWithOptions(
		RaftWithStdFSM(server.store),
		RaftWithBoltLogStore(filepath.Join(cfg.Node.DataDir, RaftLog)),
		RaftWithStdStableStore(server.store),
		RaftWithFileSnapshotStore(cfg.Node.DataDir, int(cfg.Node.MaxSnapshots), os.Stderr),
		RaftWithTCPTransport(cfg.Node.ListenPeerAddr, 32, 3*time.Second, os.Stderr),
		RaftWithConfig(func() *raft.Config {
			c := raft.DefaultConfig()
			c.LocalID = raft.ServerID(cfg.Node.ID)
			c.LogLevel = "INFO"
			return c
		}()),
		func() RaftServerOption {
			if cfg.Env().FirstRun() {
				return RaftWithBootstrap()
			}
			return RaftWithEmpty()
		}(),
		RaftWithClusters(func() []raft.Server {
			if !cfg.Env().FirstRun() {
				return nil
			}
			cluster := make([]raft.Server, 0, len(cfg.Cluster.Bootstrap))
			for _, node := range cfg.Cluster.Bootstrap {
				cluster = append(cluster, raft.Server{
					Suffrage: raft.Voter,
					ID:       raft.ServerID(node.Name),
					Address:  raft.ServerAddress(node.PeerAddr),
				})
			}
			return cluster
		}()),
	); err != nil {
		return nil, errors.Wrap(err, "NewServer")
	}

	// init distributed lock backend
	if server.lockerBackend, err = dlocker.NewLockerBackend(server.store, server.applyLog); err != nil {
		return nil, errors.Wrap(err, "NewServer")
	}

	// init requests merged
	if cfg.Node.RequestsMerged {
		server.logApplyer = NewRaftMultipleLogApply(
			context.Background(),
			64,
			300*time.Millisecond,
			3*time.Second,
			server.raft.Apply,
		)
	} else {
		server.logApplyer = NewRaftSingeLogApply(server.raft.Apply)
	}

	// start daemon service
	go server._stateUpdater()

	return server, nil
}

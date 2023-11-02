package red

import (
	"bytes"
	"context"
	"github.com/pkg/errors"

	"net"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"google.golang.org/grpc"

	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/config"
	"github.com/RealFax/RedQueen/locker"
	"github.com/RealFax/RedQueen/store"
)

var (
	bufferPool = sync.Pool{New: func() any {
		return &bytes.Buffer{}
	}}
)

type Server struct {
	term      uint64 // [ATOMIC]
	clusterID string

	cfg *config.Config

	store         store.Store
	lockerBackend locker.Backend
	logApplyer    RaftApply

	raft        *Raft
	grpcServer  *grpc.Server
	pprofServer *pprofServer

	stateNotify sync.Map // map[string]chan bool

	serverpb.UnimplementedKVServer
	serverpb.UnimplementedLockerServer
	serverpb.UnimplementedRedQueenServer
}

func (s *Server) namespace(namespace *string) (store.Namespace, error) {
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
		server = Server{
			clusterID: cfg.Node.ID,
			cfg:       cfg,
		}
		err error
	)

	if cfg.Misc.PPROF {
		server.pprofServer, err = newPprofServer()
		if err != nil {
			return nil, errors.Wrap(err, "pprof server")
		}
		go server.pprofServer.Run()
	}

	// init server store backend
	if server.store, err = newStoreBackend(cfg.Store, cfg.Node.DataDir); err != nil {
		return nil, err
	}

	// init server grpc
	server.grpcServer = grpc.NewServer()
	serverpb.RegisterKVServer(server.grpcServer, &server)
	serverpb.RegisterLockerServer(server.grpcServer, &server)
	serverpb.RegisterRedQueenServer(server.grpcServer, &server)

	// init server raft
	if server.raft, err = NewRaft(RaftConfig{
		Bootstrap:    cfg.Env().FirstRun(),
		MaxSnapshots: int(cfg.Node.MaxSnapshots),
		ServerID:     cfg.Node.ID,
		Addr:         cfg.Node.ListenPeerAddr,
		DataDir:      cfg.Node.DataDir,
		Store:        server.store,
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
		return nil, errors.Wrap(err, "NewServer")
	}

	// init distributed lock backend
	if server.lockerBackend, err = NewLockerBackend(server.store, server.applyLog); err != nil {
		return nil, errors.Wrap(err, "NewServer")
	}

	// init requests merged
	if cfg.Node.RequestsMerged {
		server.logApplyer = NewRaftMultipleLogApply(
			context.Background(),
			64,
			time.Millisecond*300,
			time.Second*3,
			server.raft.Apply,
		)
	} else {
		server.logApplyer = NewRaftSingeLogApply(server.raft.Apply)
	}

	// start daemon service
	go server._stateUpdater()

	return &server, nil
}

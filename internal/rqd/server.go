package rqd

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509/pkix"
	"github.com/RealFax/RedQueen/internal/rqd/config"
	"github.com/RealFax/RedQueen/internal/rqd/store"
	"github.com/RealFax/RedQueen/internal/version"
	"github.com/RealFax/RedQueen/pkg/dlocker"
	"github.com/RealFax/RedQueen/pkg/expr"
	"github.com/RealFax/RedQueen/pkg/grpcutil"
	"github.com/RealFax/RedQueen/pkg/httputil"
	"github.com/RealFax/RedQueen/pkg/tlsutil"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"

	"github.com/RealFax/RedQueen/api/serverpb"
)

var (
	bufferPool = sync.Pool{New: func() any {
		return &bytes.Buffer{}
	}}
)

type Server struct {
	close atomic.Bool

	ctx    context.Context
	cancel context.CancelCauseFunc

	cfg       *config.Config
	tlsConfig *tls.Config

	store         store.Store
	lockerBackend dlocker.Backend
	logApplyer    RaftApply

	raft        *Raft
	grpcServer  *grpc.Server
	httpServer  *http.Server
	pprofServer *pprofServer

	stateNotify sync.Map // map[string]chan bool

	clusterID string
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

func (s *Server) stateUpdater() {
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

func (s *Server) initTLS() (err error) {
	var cert tls.Certificate
	switch {
	case s.cfg.Node.TLS.CertFile != "" && s.cfg.Node.TLS.KeyFile != "":
		cert, err = tls.LoadX509KeyPair(s.cfg.Node.TLS.CertFile, s.cfg.Node.TLS.KeyFile)
	case s.cfg.Node.TLS.Auto:
		cert, err = tlsutil.GenX509KeyPair(pkix.Name{
			Country:            []string{"Earth"},
			Organization:       []string{"RealFax"},
			OrganizationalUnit: []string{"RedQueen"},
			CommonName:         "*",
		})
	}
	if err != nil {
		return
	}

	s.tlsConfig = &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: expr.If(s.cfg.Node.TLS.CertFile != "" && s.cfg.Node.TLS.KeyFile != "", false, true),
		NextProtos:         []string{"http/1.1", "http/2.0"},
	}
	return nil
}

func (s *Server) newNetListener(network, addr string) (net.Listener, error) {
	if s.tlsConfig != nil {
		return tls.Listen(network, addr, s.tlsConfig)
	}
	return net.Listen(network, addr)
}

func (s *Server) registerRPCServer() {
	opts := make([]grpc.ServerOption, 0, 8)
	if s.tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(s.tlsConfig)))
	}

	if s.cfg.BasicAuth != nil && len(s.cfg.BasicAuth) != 0 {
		auth := grpcutil.NewBasicAuth(grpcutil.NewMemoryBasicAuthFunc(s.cfg.BasicAuth))
		opts = append(opts, grpc.UnaryInterceptor(auth.Unary), grpc.StreamInterceptor(auth.Stream))
	}

	if s.grpcServer == nil {
		s.grpcServer = grpc.NewServer(opts...)
	}

	rpcServer := &v1RPCServer{Server: s}
	serverpb.RegisterKVServer(s.grpcServer, rpcServer)
	serverpb.RegisterLockerServer(s.grpcServer, rpcServer)
	serverpb.RegisterRedQueenServer(s.grpcServer, rpcServer)
}

func (s *Server) registerHttpServer() {
	if s.httpServer == nil {
		s.httpServer = &http.Server{
			Addr:      s.cfg.Node.ListenHttpAddr,
			TLSConfig: s.tlsConfig,
			ErrorLog:  log.New(io.Discard, "", 0),
		}
	}

	httpHandlers := &v1HttpServer{Server: s}
	router := httprouter.New()
	router.NotFound = httputil.WrapE(func(w http.ResponseWriter, r *http.Request) error {
		return httputil.NewStatus(http.StatusNotFound, 0, "Not Found")
	})
	router.PanicHandler = func(w http.ResponseWriter, _ *http.Request, _ any) {
		httputil.Any(http.StatusServiceUnavailable, 0).Message("Service Unavailable").Ok(w)
	}

	router.Handler(http.MethodGet, "/", httputil.WrapE(httpHandlers.Stats))
	router.Handler(http.MethodPost, "/lock", httputil.WrapE(httpHandlers.Lock))
	router.Handler(http.MethodDelete, "/lock", httputil.WrapE(httpHandlers.Unlock))
	router.Handler(http.MethodPatch, "/lock", httputil.WrapE(httpHandlers.TryLock))
	router.Handler(http.MethodPost, "/raft/add", httputil.WrapE(httpHandlers.AppendCluster))

	// ---- action handlers ----
	router.Handler(http.MethodPut, "/action/:bucket", httputil.WrapE(httpHandlers.Set))
	router.Handler(http.MethodGet, "/action/:bucket", httputil.WrapE(httpHandlers.Get))
	router.Handler(http.MethodDelete, "/action/:bucket", httputil.WrapE(httpHandlers.Delete))
	router.Handler(http.MethodGet, "/action/:bucket/scan", httputil.WrapE(httpHandlers.PrefixScan))
	router.Handler(http.MethodPut, "/action/:bucket/try", httputil.WrapE(httpHandlers.TrySet))

	s.httpServer.Handler = httputil.UseMiddleware(router, func(w http.ResponseWriter, r *http.Request) bool {
		w.Header().Add("Server", version.String())
		return true
	})

	if s.cfg.BasicAuth != nil && len(s.cfg.BasicAuth) != 0 {
		// use basic-auth
		s.httpServer.Handler = httputil.NewBasicAuth(
			s.httpServer.Handler,
			httputil.NewMemoryBasicAuthFunc(s.cfg.BasicAuth),
		)
	}

	if s.tlsConfig != nil {
		s.httpServer.TLSConfig = s.tlsConfig
		s.httpServer.ErrorLog = nil
	}
}

func (s *Server) ListenAndServe() error {
	if s.cfg.Node.ListenHttpAddr != "" {
		if s.httpServer == nil {
			return errors.New("ListenAndServe: http server did not complete initialization")
		}

		listener, err := s.newNetListener("tcp", s.cfg.Node.ListenHttpAddr)
		if err != nil {
			return err
		}
		go func() {
			_ = s.httpServer.Serve(listener)
		}()
	}

	// grpc server
	listener, err := net.Listen("tcp", s.cfg.Node.ListenClientAddr)
	if err != nil {
		return err
	}

	go func() {
		_ = s.grpcServer.Serve(listener)
	}()
	return nil
}

func (s *Server) Shutdown() {
	if !s.close.CompareAndSwap(false, true) {
		return
	}

	s.cancel(errors.New("server close"))

	s.raft.Shutdown()
	s.grpcServer.Stop()

	if s.cfg.Node.ListenHttpAddr != "" && s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
		defer cancel()
		s.httpServer.Shutdown(ctx)
	}

	if s.cfg.PPROF {
		s.pprofServer.Close()
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
			return nil, errors.Wrap(err, "NewServer")
		}
		go server.pprofServer.Run()
	}

	// init server actions backend
	if server.store, err = newStoreBackend(cfg.Store, cfg.Node.DataDir); err != nil {
		return nil, errors.Wrap(err, "NewServer")
	}

	// try init tls config
	if err = server.initTLS(); err != nil {
		return nil, errors.Wrap(err, "NewServer")
	}

	// init server grpc
	server.registerRPCServer()

	// init server http
	server.registerHttpServer()

	// init server raft
	if server.raft, err = NewRaftWithOptions(
		RaftWithContext(server.ctx),
		RaftWithStdFSM(server.store),
		RaftWithBoltLogStore(filepath.Join(cfg.Node.DataDir, RaftLog)),
		RaftWithStdStableStore(server.store),
		RaftWithFileSnapshotStore(cfg.Node.DataDir, int(cfg.Node.MaxSnapshots), os.Stderr),
		RaftWithTCPTransport(cfg.Node.ListenPeerAddr, 32, 10*time.Second, os.Stderr),
		RaftWithConfig(func() *raft.Config {
			c := raft.DefaultConfig()
			c.LocalID = raft.ServerID(cfg.Node.ID)
			c.LogLevel = "INFO"
			c.Logger = hclog.New(&hclog.LoggerOptions{
				Name:            "rqd",
				Level:           hclog.Info,
				Output:          os.Stderr,
				IncludeLocation: false,
				TimeFormat:      time.RFC3339,
				TimeFn:          time.Now,
				Color:           hclog.AutoColor,
				ColorHeaderOnly: true,
			})
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
	go server.stateUpdater()

	return server, nil
}

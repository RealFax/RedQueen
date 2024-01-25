package rqd

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
)

type pprofServer struct {
	listener net.Listener
	server   http.Server
}

func (s *pprofServer) Run() error {
	return s.server.Serve(s.listener)
}

func (s *pprofServer) Close() error {
	defer s.listener.Close()
	return s.server.Shutdown(context.Background())
}

func newPprofServer() (*pprofServer, error) {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, err
	}
	log.Println("pprof server address: ", listener.Addr())

	mux := http.NewServeMux()
	mux.HandleFunc("/pprof/", pprof.Index)
	mux.HandleFunc("/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/pprof/trace", pprof.Trace)
	return &pprofServer{
		listener: listener,
		server: http.Server{
			Handler: mux,
		},
	}, nil
}

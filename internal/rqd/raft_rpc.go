package rqd

import (
	"crypto/tls"
	"crypto/x509/pkix"
	"github.com/RealFax/RedQueen/pkg/tlsutil"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"io"
	"net"
	"net/netip"
	"time"
)

type RaftTLStreamLayer struct {
	autoTLS  bool
	config   *tls.Config
	addr     net.Addr
	listener net.Listener
}

func (r RaftTLStreamLayer) Accept() (net.Conn, error) { return r.listener.Accept() }
func (r RaftTLStreamLayer) Close() error              { return r.listener.Close() }
func (r RaftTLStreamLayer) Addr() net.Addr            { return r.listener.Addr() }
func (r RaftTLStreamLayer) Dial(addr raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	return tls.DialWithDialer(&net.Dialer{Timeout: timeout}, "tcp", string(addr), r.config)
}

func newTLSTransport(
	autoTLS bool,
	addr string,
	advertise net.Addr,
	cfg *tls.Config,
	transportCreator func(layer raft.StreamLayer) *raft.NetworkTransport,
) (*raft.NetworkTransport, error) {
	inet, err := netip.ParseAddrPort(addr)
	if err != nil {
		return nil, err
	}

	switch {
	case err != nil:
		return nil, err
	case !inet.IsValid():
		return nil, errors.New("invalid transport listener address")
	case inet.Addr().IsUnspecified():
		return nil, errors.New("local bind address is not advertisable")
	}

	listener, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		return nil, err
	}

	if autoTLS {
		cfg.InsecureSkipVerify = true
	}

	return transportCreator(&RaftTLStreamLayer{
		autoTLS:  autoTLS,
		config:   cfg,
		addr:     advertise,
		listener: listener,
	}), nil
}

func NewTLSTransportWithGenerator(
	addr string,
	advertise net.Addr,
	config *raft.NetworkTransportConfig,
) (*raft.NetworkTransport, error) {
	cert, err := tlsutil.GenX509KeyPair(pkix.Name{
		CommonName:         addr,
		Country:            []string{"Earth"},
		Organization:       []string{"RealFax"},
		OrganizationalUnit: []string{"RedQueen"},
	})
	if err != nil {
		return nil, err
	}

	return newTLSTransport(true, addr, advertise, &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}, func(layer raft.StreamLayer) *raft.NetworkTransport {
		config.Stream = layer
		return raft.NewNetworkTransportWithConfig(config)
	})
}

func NewTLSTransport(
	addr string,
	advertise net.Addr,
	maxPool int,
	timeout time.Duration,
	logOutput io.Writer,
	cfg *tls.Config,
) (*raft.NetworkTransport, error) {
	return newTLSTransport(false, addr, advertise, cfg, func(layer raft.StreamLayer) *raft.NetworkTransport {
		return raft.NewNetworkTransport(layer, maxPool, timeout, logOutput)
	})
}

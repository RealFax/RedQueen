package rqd_test

import (
	red "github.com/RealFax/RedQueen/internal/rqd"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/require"
	"net"
	"testing"
)

func TestNewTLSTransportWithGenerator(t *testing.T) {
	addr := "127.0.0.1:1234"
	advertise := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 5678,
	}
	config := &raft.NetworkTransportConfig{}
	transport, err := red.NewTLSTransportWithGenerator(addr, advertise, config)
	require.NoError(t, err)
	require.NotNil(t, transport)
}

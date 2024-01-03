package client

import (
	"context"
	"github.com/RealFax/RedQueen/pkg/balancer"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"sync/atomic"
)

type BalancerConn struct {
	ID string
	*grpc.ClientConn
}

func (c *BalancerConn) Key() string             { return c.ID }
func (c *BalancerConn) Value() *grpc.ClientConn { return c.ClientConn }

type ConnectionManager struct {
	state    atomic.Bool
	ctx      context.Context
	endpoint string
	balancer balancer.LoadBalance[string, *grpc.ClientConn]
}

func (m *ConnectionManager) Target() string { return m.endpoint }

func (m *ConnectionManager) Alloc() (*grpc.ClientConn, error) {
	conn, err := m.balancer.Next()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (m *ConnectionManager) Close() error {
	if !m.state.Load() {
		return errors.New("connection manager close")
	}
	m.state.Store(true)

	for i := 0; i < int(m.balancer.Size()); i++ {
		conn, err := m.balancer.Next()
		if err != nil {
			return err
		}
		_ = conn.Close()
	}

	return nil
}

func NewConnectionManager(ctx context.Context, endpoint string, maxConn int, opts ...grpc.DialOption) (*ConnectionManager, error) {
	manager := &ConnectionManager{
		ctx:      ctx,
		endpoint: endpoint,
		balancer: balancer.NewRoundRobin[string, *grpc.ClientConn](),
	}

	for i := 0; i < maxConn; i++ {
		conn, err := grpc.DialContext(
			ctx,
			endpoint,
			opts...,
		)
		if err != nil {
			return nil, err
		}

		manager.balancer.Append(&BalancerConn{
			ID:         uuid.NewString(),
			ClientConn: conn,
		})
	}

	return manager, nil
}

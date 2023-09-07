package client

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"sync/atomic"
)

type GrpcPoolConn struct {
	state   atomic.Int32
	release func(c *GrpcPoolConn)
	ID      string
	*grpc.ClientConn
}

func (c *GrpcPoolConn) Close() error {
	c.state.Store(-1)
	return c.ClientConn.Close()
}

func (c *GrpcPoolConn) Release() {
	c.release(c)
}

type GrpcPool struct {
	state    atomic.Int32
	ctx      context.Context
	endpoint string

	opts  []grpc.DialOption
	rings chan *GrpcPoolConn
}

func (p *GrpcPool) factory() (*GrpcPoolConn, error) {
	conn, err := grpc.DialContext(
		p.ctx,
		p.endpoint,
		p.opts...,
	)
	if err != nil {
		return nil, err
	}
	return &GrpcPoolConn{
		state:      atomic.Int32{},
		release:    p.Free,
		ID:         uuid.New().String(),
		ClientConn: conn,
	}, nil
}

func (p *GrpcPool) Target() string {
	return p.endpoint
}

func (p *GrpcPool) Alloc() (*GrpcPoolConn, error) {
	if p.state.Load() == -1 {
		return nil, errors.New("conn pool closed")
	}

alloc:
	select {
	case c := <-p.rings:
		if c == nil {
			return nil, errors.New("conn pool closed")
		}

		if c.state.Load() == -1 {
			goto alloc
		}

		return c, nil
	default:
		return p.factory()
	}
}

func (p *GrpcPool) Free(c *GrpcPoolConn) {
	if c.state.Load() == -1 {
		return
	}

	select {
	case p.rings <- c:
		return
	default:
		_ = c.Close()
	}
}

func (p *GrpcPool) Close() error {
	if p.state.Load() == -1 {
		return errors.New("conn pool closed")
	}
	p.state.Store(-1)

	for {
		select {
		case c := <-p.rings:
			_ = c.Close()
		default:
			return nil
		}
	}
}

func NewGrpcPool(ctx context.Context, endpoint string, psize int, opts ...grpc.DialOption) (*GrpcPool, error) {
	pool := &GrpcPool{
		ctx:      ctx,
		endpoint: endpoint,
		rings:    make(chan *GrpcPoolConn, psize),
		opts:     opts,
	}

	for i := 0; i < psize; i++ {
		factory, err := pool.factory()
		if err != nil {
			return nil, err
		}
		pool.rings <- factory
	}
	return pool, nil
}

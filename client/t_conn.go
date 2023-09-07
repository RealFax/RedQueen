package client

import (
	"context"
	"crypto/rand"
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"math/big"
	"sync"
	"sync/atomic"
)

type Conn interface {
	ReadOnly() (*GrpcPoolConn, error)
	WriteOnly() (*GrpcPoolConn, error)
	Close() error
}

type clientConn struct {
	state     atomic.Bool
	mu        sync.Mutex
	ctx       context.Context
	writeOnly *GrpcPool
	readOnly  map[string]*GrpcPool

	endpoints []string
}

func (c *clientConn) swapLeaderConn(new string) error {
	c.mu.Lock()

	if c.writeOnly != nil {
		c.readOnly[c.writeOnly.Target()] = c.writeOnly
	}

	nextConn, ok := c.readOnly[new]
	if !ok {
		c.mu.Unlock()
		return errors.New("target conn not found")
	}

	c.writeOnly = nextConn
	c.mu.Unlock()
	return nil
}

func (c *clientConn) listenLeader() {
	wg := sync.WaitGroup{}
	wg.Add(len(c.readOnly))

	finalTry := func(conn *GrpcPoolConn) {
		defer conn.Release()
		var (
			err     error
			monitor serverpb.RedQueen_LeaderMonitorClient
			resp    *serverpb.LeaderMonitorResponse
			call    = serverpb.NewRedQueenClient(conn)
		)

		// make a preliminary check of to conn
		xResp, xErr := call.RaftState(c.ctx, &serverpb.RaftStateRequest{})
		if xErr == nil {
			if xResp.State == serverpb.RaftState_leader {
				_ = c.swapLeaderConn(conn.Target())
			}
			wg.Done()
		}

		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				if monitor, err = call.LeaderMonitor(c.ctx, &serverpb.LeaderMonitorRequest{}); err != nil {
					continue
				}
				for {
					if resp, err = monitor.Recv(); err != nil {
						continue
					}
					if resp.Leader {
						_ = c.swapLeaderConn(conn.Target())
					}
				}
			}
		}
	}

	for _, conn := range c.readOnly {
		cc, err := conn.Alloc()
		if err != nil {
			panic(err)
		}
		go finalTry(cc)
	}

	wg.Wait()
}

func (c *clientConn) ReadOnly() (*GrpcPoolConn, error) {
	size := len(c.readOnly)
	if size == 0 {
		return nil, errors.New("read-only not maintained")
	}

	var (
		step     int64
		round, _ = rand.Int(rand.Reader, big.NewInt(int64(size)))
	)

	for _, pool := range c.readOnly {
		step++
		if step != round.Int64() {
			continue
		}

		cc, err := pool.Alloc()
		if err != nil {
			return nil, err
		}
		return cc, nil
	}

	return nil, errors.New("unexpected")
}

func (c *clientConn) WriteOnly() (*GrpcPoolConn, error) {
	if c.writeOnly == nil {
		return nil, errors.New("write-only not maintained")
	}
	cc, err := c.writeOnly.Alloc()
	if err != nil {
		return nil, err
	}
	return cc, nil
}

func (c *clientConn) Close() error {
	if !c.state.Load() {
		return errors.New("client connect has closed")
	}
	c.state.Store(false)

	if c.writeOnly != nil {
		c.writeOnly.Close()
		c.writeOnly = nil
	}

	for key, conn := range c.readOnly {
		conn.Close()
		delete(c.readOnly, key)
	}

	return nil
}

func NewClientConn(ctx context.Context, endpoints []string, opts ...grpc.DialOption) (Conn, error) {
	cc := &clientConn{
		state:     atomic.Bool{},
		ctx:       ctx,
		writeOnly: nil,
		readOnly:  make(map[string]*GrpcPool),
		endpoints: endpoints,
	}
	cc.state.Store(true)

	var (
		err  error
		pool *GrpcPool
	)

	// init
	for _, endpoint := range endpoints {
		if pool, err = NewGrpcPool(ctx, endpoint, 16, opts...); err != nil {
			return nil, err
		}
		cc.readOnly[endpoint] = pool
	}

	// start listen
	cc.listenLeader()

	return cc, nil
}

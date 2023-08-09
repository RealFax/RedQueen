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
	ReadOnly() (*grpc.ClientConn, error)
	WriteOnly() (*grpc.ClientConn, error)
	Close() error
}

type clientConn struct {
	state     atomic.Bool
	mu        sync.Mutex
	ctx       context.Context
	writeOnly *grpc.ClientConn
	readOnly  map[string]*grpc.ClientConn

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

	finalTry := func(conn *grpc.ClientConn) {
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
		go finalTry(conn)
	}

	wg.Wait()
}

func (c *clientConn) ReadOnly() (*grpc.ClientConn, error) {
	size := len(c.readOnly)
	if size == 0 {
		return nil, errors.New("read-only not maintained")
	}

	var (
		step     int64
		round, _ = rand.Int(rand.Reader, big.NewInt(int64(size)))
	)

	for _, conn := range c.readOnly {
		step++
		if step == round.Int64() {
			return conn, nil
		}
	}

	return nil, errors.New("unexpected")
}

func (c *clientConn) WriteOnly() (*grpc.ClientConn, error) {
	if c.writeOnly == nil {
		return nil, errors.New("write-only not maintained")
	}
	return c.writeOnly, nil
}

func (c *clientConn) Close() error {
	if !c.state.Load() {
		return errors.New("client connect has closed")
	}
	c.state.Store(false)

	c.writeOnly.Close()
	c.writeOnly = nil

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
		readOnly:  make(map[string]*grpc.ClientConn),
		endpoints: endpoints,
	}
	cc.state.Store(true)

	var (
		err  error
		conn *grpc.ClientConn
	)

	// init
	for _, endpoint := range endpoints {
		if conn, err = grpc.DialContext(ctx, endpoint, opts...); err != nil {
			return nil, err
		}
		cc.readOnly[endpoint] = conn
	}

	// start listen
	cc.listenLeader()

	return cc, nil
}

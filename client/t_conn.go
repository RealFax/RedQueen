package client

import (
	"context"
	"crypto/rand"
	"math/big"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/RealFax/RedQueen/api/serverpb"
)

type Conn interface {
	ReadOnly() (*grpc.ClientConn, error)
	WriteOnly() (*grpc.ClientConn, error)
}

type clientConn struct {
	ctx       context.Context
	mu        sync.Mutex
	writeOnly *grpc.ClientConn
	readOnly  map[string]*grpc.ClientConn

	endpoints []string
}

func (c *clientConn) swapLeaderConn(new string) error {
	c.mu.Lock()
	c.readOnly[c.writeOnly.Target()] = c.writeOnly

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
	finalTry := func(ch chan bool, conn *grpc.ClientConn) {
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
					ch <- resp.Leader
				}
			}
		}
	}
	swapLeader := func(ch chan bool, conn *grpc.ClientConn) {
		for {
			select {
			case <-c.ctx.Done():
				return
			case state := <-ch:
				if !state { // next round
					continue
				}
				_ = c.swapLeaderConn(conn.Target())
			}
		}
	}

	for _, conn := range c.readOnly {
		ch := make(chan bool, 1)
		go finalTry(ch, conn)
		go swapLeader(ch, conn)
	}
}

func (c *clientConn) ReadOnly() (*grpc.ClientConn, error) {
	size := len(c.readOnly)
	if size == 0 {
		return nil, errors.New("read-only not maintained")
	}

	var (
		step     int64 = 1
		round, _       = rand.Int(rand.Reader, big.NewInt(int64(size)))
	)

	for _, conn := range c.readOnly {
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

func newClientConn(ctx context.Context, endpoints []string, syncConn bool) (Conn, error) {
	cc := &clientConn{
		ctx:       ctx,
		writeOnly: nil,
		readOnly:  make(map[string]*grpc.ClientConn),
		endpoints: endpoints,
	}

	var (
		err  error
		conn *grpc.ClientConn
		opts = []grpc.DialOption{
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                time.Second * 3,
				Timeout:             time.Millisecond * 100,
				PermitWithoutStream: true,
			}),
		}
	)

	if syncConn {
		opts = append(opts, grpc.WithBlock())
	}

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

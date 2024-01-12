package client

import (
	"context"
	"crypto/rand"
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"math/big"
	"sync"
	"sync/atomic"
	"time"
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
	writeOnly *ConnectionManager
	readOnly  map[string]*ConnectionManager

	endpoints []string
}

func (c *clientConn) swapLeader(new string) error {
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

func (c *clientConn) leaderWatcher() {
	wg := sync.WaitGroup{}
	wg.Add(len(c.readOnly))

	leaderIter := func(conn *grpc.ClientConn) {
		var (
			monitor serverpb.RedQueen_LeaderMonitorClient
			mAck    *serverpb.LeaderMonitorResponse
			call    = serverpb.NewRedQueenClient(conn)
		)

		// make a preliminary check of to conn
		ack, err := call.RaftState(c.ctx, &emptypb.Empty{})
		if err == nil && ack.State == serverpb.RaftState_leader {
			_ = c.swapLeader(conn.Target())
		}

		wg.Done() // wait group have done first

		// ---- leader long watcher ----
		for {
			select {
			case <-c.ctx.Done():
				return
			default:
				if monitor, err = call.LeaderMonitor(c.ctx, &serverpb.LeaderMonitorRequest{}); err != nil {
					time.Sleep(time.Millisecond * 100)
					continue
				}
				for {
					if mAck, err = monitor.Recv(); err != nil {
						break
					}
					if mAck.Leader {
						_ = c.swapLeader(conn.Target())
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
		go leaderIter(cc)
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

func (c *clientConn) WriteOnly() (*grpc.ClientConn, error) {
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
		_ = c.writeOnly.Close()
		c.writeOnly = nil
	}

	for key, conn := range c.readOnly {
		_ = conn.Close()
		delete(c.readOnly, key)
	}

	return nil
}

func NewClientConn(ctx context.Context, endpoints []string, opts ...grpc.DialOption) (Conn, error) {
	cc := &clientConn{
		state:     atomic.Bool{},
		ctx:       ctx,
		writeOnly: nil,
		readOnly:  make(map[string]*ConnectionManager),
		endpoints: endpoints,
	}
	cc.state.Store(true)

	var (
		err     error
		manager *ConnectionManager
	)

	// init
	for _, endpoint := range endpoints {
		if manager, err = NewConnectionManager(
			ctx,
			endpoint,
			int(atomic.LoadInt64(&maxOpenConn)+1), // include leader watcher alloc
			opts...,
		); err != nil {
			return nil, err
		}
		cc.readOnly[endpoint] = manager
	}

	// start listen
	cc.leaderWatcher()

	return cc, nil
}

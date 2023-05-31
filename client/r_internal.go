package client

import (
	"context"
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/pkg/errors"
)

type InternalClient interface {
	AppendCluster(context.Context, *serverpb.AppendClusterRequest) error
	LeaderMonitor(context.Context, *chan bool) error
}

type internalClient struct {
	ctx  context.Context
	conn Conn
}

func (c *internalClient) AppendCluster(ctx context.Context, in *serverpb.AppendClusterRequest) error {
	client, err := newClientCall[serverpb.RedQueenClient](true, c.conn, serverpb.NewRedQueenClient)
	if err != nil {
		return err
	}

	_, err = client.AppendCluster(ctx, in)
	return err
}

func (c *internalClient) LeaderMonitor(ctx context.Context, ch *chan bool) error {
	if len(*ch) != 0 || cap(*ch) != 1 {
		return errors.New("invalid receiver channel")
	}

	client, err := newClientCall[serverpb.RedQueenClient](false, c.conn, serverpb.NewRedQueenClient)
	if err != nil {
		return err
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	monitor, err := client.LeaderMonitor(ctx, &serverpb.LeaderMonitorRequest{})
	if err != nil {
		return err
	}
	var resp *serverpb.LeaderMonitorResponse

	for {
		if resp, err = monitor.Recv(); err != nil {
			return err
		}
		*ch <- resp.Leader
	}
}

func newInternalClient(ctx context.Context, conn Conn) InternalClient {
	return &internalClient{
		ctx:  ctx,
		conn: conn,
	}
}

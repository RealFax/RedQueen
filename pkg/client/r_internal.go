package client

import (
	"context"
	"github.com/pkg/errors"

	"github.com/RealFax/RedQueen/api/serverpb"
)

type InternalClient interface {
	AppendCluster(ctx context.Context, serverID string, peerAddr string, voter bool) error
	LeaderMonitor(ctx context.Context, recv *chan bool) error
	Snapshot(ctx context.Context, serverPath *string) error
}

type internalClient struct {
	ctx  context.Context
	conn Conn
}

func (c *internalClient) AppendCluster(ctx context.Context, serverID, peerAddr string, voter bool) error {
	client, err := newClientCall[serverpb.RedQueenClient](true, c.conn, serverpb.NewRedQueenClient)
	if err != nil {
		return err
	}

	_, err = client.instance.AppendCluster(ctx, &serverpb.AppendClusterRequest{
		ServerId: serverID,
		PeerAddr: peerAddr,
		Voter:    voter,
	})
	return err
}

func (c *internalClient) LeaderMonitor(ctx context.Context, recv *chan bool) error {
	if len(*recv) != 0 || cap(*recv) != 1 {
		return errors.New("invalid receiver channel")
	}

	client, err := newClientCall[serverpb.RedQueenClient](false, c.conn, serverpb.NewRedQueenClient)
	if err != nil {
		return err
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	monitor, err := client.instance.LeaderMonitor(ctx, &serverpb.LeaderMonitorRequest{})
	if err != nil {
		return err
	}
	var resp *serverpb.LeaderMonitorResponse

	for {
		if resp, err = monitor.Recv(); err != nil {
			return err
		}
		*recv <- resp.Leader
	}
}

func (c *internalClient) Snapshot(ctx context.Context, serverPath *string) error {
	client, err := newClientCall(false, c.conn, serverpb.NewRedQueenClient)
	if err != nil {
		return err
	}

	if _, err = client.instance.RaftSnapshot(ctx, &serverpb.RaftSnapshotRequest{
		Path: serverPath,
	}); err != nil {
		return err
	}

	return nil
}

func newInternalClient(ctx context.Context, conn Conn) InternalClient {
	return &internalClient{
		ctx:  ctx,
		conn: conn,
	}
}

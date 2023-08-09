package client

import (
	"context"
	"google.golang.org/grpc"
)

type Client struct {
	InternalClient
	KvClient
	LockerClient

	conn       Conn
	ctx        context.Context
	cancelFunc context.CancelFunc
	endpoints  []string
}

func (c *Client) Close() error {
	c.cancelFunc()
	return c.conn.Close()
}

func New(ctx context.Context, endpoints []string, opts ...grpc.DialOption) (*Client, error) {
	var (
		err    error
		client = &Client{
			endpoints: endpoints,
		}
	)

	client.ctx, client.cancelFunc = context.WithCancel(ctx)

	if client.conn, err = NewClientConn(ctx, endpoints, opts...); err != nil {
		return nil, err
	}

	client.InternalClient = newInternalClient(ctx, client.conn)
	client.KvClient = newKvClient(ctx, client.conn)
	client.LockerClient = newLockerClient(ctx, client.conn)

	return client, nil
}

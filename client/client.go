package client

import (
	"context"
	"google.golang.org/grpc"
)

type Client struct {
	InternalClient
	KvClient
	LockerClient

	conn      Conn
	ctx       context.Context
	endpoints []string
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func New(ctx context.Context, endpoints []string, opts ...grpc.DialOption) (*Client, error) {
	var (
		err    error
		client = &Client{
			endpoints: endpoints,
		}
	)

	if client.conn, err = NewClientConn(ctx, endpoints, opts...); err != nil {
		return nil, err
	}

	client.InternalClient = newInternalClient(ctx, client.conn)
	client.KvClient = newKvClient(ctx, client.conn)
	client.LockerClient = newLockerClient(ctx, client.conn)

	return client, nil
}

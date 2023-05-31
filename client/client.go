package client

import "context"

type Client struct {
	InternalClient
	KvClient
	LockerClient

	conn      Conn
	ctx       context.Context
	endpoints []string
}

func New(ctx context.Context, endpoints []string) (*Client, error) {
	var (
		err    error
		client = &Client{
			ctx:       ctx,
			endpoints: endpoints,
		}
	)

	if client.conn, err = newClientConn(ctx, endpoints, false); err != nil {
		return nil, err
	}

	client.InternalClient = newInternalClient(ctx, client.conn)
	client.KvClient = newKvClient(ctx, client.conn)
	client.LockerClient = newLockerClient(ctx, client.conn)

	return client, nil
}

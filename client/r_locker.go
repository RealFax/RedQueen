package client

import (
	"context"
	"github.com/RealFax/RedQueen/api/serverpb"
)

type LockerClient interface {
	Lock(ctx context.Context, lockID string, ttl int64) error
	Unlock(ctx context.Context, lockID string) error
	TryLock(ctx context.Context, lockID string, ttl, deadline int64) error
}

type lockerClient struct {
	ctx  context.Context
	conn Conn
}

func (c *lockerClient) Lock(ctx context.Context, lockID string, ttl int64) error {
	client, err := newClientCall[serverpb.LockerClient](true, c.conn, serverpb.NewLockerClient)
	if err != nil {
		return err
	}

	_, err = client.Lock(ctx, &serverpb.LockRequest{
		LockId: lockID,
		Ttl:    ttl,
	})
	return err
}

func (c *lockerClient) Unlock(ctx context.Context, lockID string) error {
	client, err := newClientCall[serverpb.LockerClient](true, c.conn, serverpb.NewLockerClient)
	if err != nil {
		return err
	}

	_, err = client.Unlock(ctx, &serverpb.UnlockRequest{
		LockId: lockID,
	})
	return err
}

func (c *lockerClient) TryLock(ctx context.Context, lockID string, ttl, deadline int64) error {
	client, err := newClientCall[serverpb.LockerClient](true, c.conn, serverpb.NewLockerClient)
	if err != nil {
		return err
	}

	_, err = client.TryLock(ctx, &serverpb.TryLockRequest{
		LockId:   lockID,
		Ttl:      ttl,
		Deadline: deadline,
	})
	return err
}

func newLockerClient(ctx context.Context, conn Conn) LockerClient {
	return &lockerClient{
		ctx:  ctx,
		conn: conn,
	}
}

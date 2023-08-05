// mutex lock is a simple wrapper (distributed-lock)

package client

import (
	"context"
	"github.com/pkg/errors"
	"time"
)

type MutexLock struct {
	// internal mutex
	ttl     int32
	ctx     context.Context
	client  LockerClient
	assetID string
}

func (l *MutexLock) AssetID(assetID string) { l.assetID = assetID }
func (l *MutexLock) Lock() error            { return l.client.Lock(l.ctx, l.assetID, l.ttl) }
func (l *MutexLock) Unlock() error          { return l.client.Unlock(l.ctx, l.assetID) }
func (l *MutexLock) TryLock(deadline time.Time) error {
	if deadline.Before(deadline) {
		return errors.New("invalid deadline")
	}
	return l.client.TryLock(l.ctx, l.assetID, l.ttl, deadline.UnixNano())
}

func NewMutexLock(ctx context.Context, client LockerClient, ttl int32, assetID string) *MutexLock {
	return &MutexLock{
		ttl:     ttl,
		ctx:     ctx,
		client:  client,
		assetID: assetID,
	}
}

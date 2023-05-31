package client

import (
	"context"
	"github.com/RealFax/RedQueen/api/serverpb"
)

type Value struct {
	Data []byte
	TTL  *uint32
}

type KvClient interface {
	Set(ctx context.Context, key, value []byte, ttl uint32, namespace *string) error
	Get(ctx context.Context, key []byte, namespace *string) (*Value, error)
	Delete(ctx context.Context, key []byte, namespace *string) error
	Watch(ctx context.Context, watcher *Watcher) error
}

type kvClient struct {
	ctx  context.Context
	conn Conn
}

func (c *kvClient) Set(ctx context.Context, key, value []byte, ttl uint32, namespace *string) error {
	client, err := newClientCall[serverpb.KVClient](true, c.conn, serverpb.NewKVClient)
	if err != nil {
		return err
	}

	_, err = client.Set(ctx, &serverpb.SetRequest{
		Key:         key,
		Value:       value,
		Ttl:         ttl,
		IgnoreValue: ignoreBytes(key),
		IgnoreTtl:   ignoreBytes(value),
		Namespace:   namespace,
	})
	return err
}

func (c *kvClient) Get(ctx context.Context, key []byte, namespace *string) (*Value, error) {
	client, err := newClientCall[serverpb.KVClient](false, c.conn, serverpb.NewKVClient)
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(ctx, &serverpb.GetRequest{
		Key:       key,
		Namespace: namespace,
	})
	if err != nil {
		return nil, err
	}
	return &Value{
		Data: resp.Value,
		TTL:  resp.Ttl,
	}, nil
}

func (c *kvClient) TrySet(ctx context.Context, key, value []byte, ttl uint32, namespace *string) error {
	client, err := newClientCall[serverpb.KVClient](true, c.conn, serverpb.NewKVClient)
	if err != nil {
		return err
	}

	_, err = client.TrySet(ctx, &serverpb.SetRequest{
		Key:         key,
		Value:       value,
		Ttl:         ttl,
		IgnoreValue: ignoreBytes(key),
		IgnoreTtl:   ignoreBytes(value),
		Namespace:   namespace,
	})
	return err
}

func (c *kvClient) Delete(ctx context.Context, key []byte, namespace *string) error {
	client, err := newClientCall[serverpb.KVClient](true, c.conn, serverpb.NewKVClient)
	if err != nil {
		return err
	}

	_, err = client.Delete(ctx, &serverpb.DeleteRequest{
		Key:       key,
		Namespace: namespace,
	})
	return err
}

func (c *kvClient) Watch(ctx context.Context, watcher *Watcher) error {
	client, err := newClientCall[serverpb.KVClient](false, c.conn, serverpb.NewKVClient)
	if err != nil {
		return err
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	watch, err := client.Watch(ctx, &serverpb.WatchRequest{
		Key:          watcher.key,
		IgnoreErrors: watcher.ignoreErrors,
		Namespace:    watcher.namespace,
		BufSize: func() *uint32 {
			if watcher.bufSize != 0 {
				return &watcher.bufSize
			}
			dup := DefaultWatchBufSize
			return &dup
		}(),
	})
	if err != nil {
		return err
	}

	defer watcher.Close()
	var resp *serverpb.WatchResponse

	for {
		if resp, err = watch.Recv(); err != nil {
			return err
		}
		if watcher.close.Load() {
			return ErrWatcherClosed
		}

		watcher.ch <- &WatchValue{
			seq:       resp.UpdateSeq,
			Timestamp: resp.Timestamp,
			Data:      resp.Data,
		}
	}

}

func newKvClient(ctx context.Context, conn Conn) KvClient {
	return &kvClient{
		ctx:  ctx,
		conn: conn,
	}
}

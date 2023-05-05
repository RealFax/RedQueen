package RedQueen

import (
	"context"
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/store"
	"github.com/pkg/errors"
	"sync/atomic"
	"time"
)

type KV interface {
	Set(context.Context, *serverpb.SetRequest) (*serverpb.SetResponse, error)
	Get(context.Context, *serverpb.GetRequest) (*serverpb.GetResponse, error)
	TrySet(context.Context, *serverpb.SetRequest) (*serverpb.SetResponse, error)
	Delete(context.Context, *serverpb.DeleteRequest) (*serverpb.DeleteResponse, error)
	Watch(*serverpb.WatchRequest, serverpb.KV_WatchServer) error
}

func (s *Server) Set(ctx context.Context, req *serverpb.SetRequest) (*serverpb.SetResponse, error) {

	if _, err := s.raftApply(ctx, time.Millisecond*500, &LogPayload{
		Command: func() Command {
			if req.IgnoreTtl {
				return Set
			}
			return SetWithTTL
		}(),
		TTL: func() *uint32 {
			return &req.Ttl
		}(),
		Namespace: func() string {
			if req.Namespace != nil {
				return *req.Namespace
			}
			return store.DefaultNamespace
		}(),
		Key: req.Key,
		Value: func() []byte {
			if req.IgnoreValue {
				return []byte{}
			}
			return req.Value
		}(),
	}); err != nil {
		return nil, err
	}

	return &serverpb.SetResponse{
		Header: &serverpb.ResponseHeader{
			ClusterId: s.clusterID,
			RaftTerm:  atomic.LoadUint64(&s.term),
		},
	}, nil
}

func (s *Server) Get(ctx context.Context, req *serverpb.GetRequest) (*serverpb.GetResponse, error) {
	storeAPI, err := s.currentNamespace(req.Namespace)
	if err != nil {
		return nil, err
	}

	value, err := storeAPI.Get(req.Key)
	if err != nil {
		return nil, err
	}

	return &serverpb.GetResponse{
		Header: &serverpb.ResponseHeader{
			ClusterId: s.clusterID,
			RaftTerm:  atomic.LoadUint64(&s.term),
		},
		Value: value.Data,
		Ttl:   nil, // unimplemented
	}, nil
}

func (s *Server) TrySet(ctx context.Context, req *serverpb.SetRequest) (*serverpb.SetResponse, error) {
	return nil, nil
}

func (s *Server) Delete(ctx context.Context, req *serverpb.DeleteRequest) (*serverpb.DeleteResponse, error) {
	return nil, nil
}

func (s *Server) Watch(req *serverpb.WatchRequest, stream serverpb.KV_WatchServer) error {

	storeAPI, err := s.currentNamespace(req.Namespace)
	if err != nil {
		return err
	}

	if !req.IgnoreErrors {
		if _, err = storeAPI.Get(req.Key); err == store.ErrKeyNotFound {
			return err
		}
	}

	watcher, err := storeAPI.Watch(req.Key)
	if err != nil {
		return err
	}
	defer watcher.Close()

	for {
		select {
		case value := <-watcher.Notify():
			if value.Deleted() && !req.IgnoreErrors {
				return errors.New("key has deleted")
			}
			if err = stream.Send(&serverpb.WatchResponse{
				Header: &serverpb.ResponseHeader{
					ClusterId: s.clusterID,
					RaftTerm:  atomic.LoadUint64(&s.term),
				},
				UpdateSeq: value.Seq,
				Timestamp: value.Timestamp,
				Data:      *value.Data,
			}); err != nil {
				// unrecoverable error
				return err
			}
		}
	}

}

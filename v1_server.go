package red

import (
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/raft"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/locker"
	"github.com/RealFax/RedQueen/store"
)

type KV interface {
	Set(context.Context, *serverpb.SetRequest) (*serverpb.SetResponse, error)
	Get(context.Context, *serverpb.GetRequest) (*serverpb.GetResponse, error)
	TrySet(context.Context, *serverpb.SetRequest) (*serverpb.SetResponse, error)
	Delete(context.Context, *serverpb.DeleteRequest) (*serverpb.DeleteResponse, error)
	Watch(*serverpb.WatchRequest, serverpb.KV_WatchServer) error
}

type Locker interface {
	Lock(context.Context, *serverpb.LockRequest) (*serverpb.LockResponse, error)
	Unlock(context.Context, *serverpb.UnlockRequest) (*serverpb.UnlockResponse, error)
	TryLock(context.Context, *serverpb.TryLockRequest) (*serverpb.TryLockResponse, error)
}

type RedQueen interface {
	AppendCluster(context.Context, *serverpb.AppendClusterRequest) (*serverpb.AppendClusterResponse, error)
	LeaderMonitor(*serverpb.LeaderMonitorRequest, serverpb.RedQueen_LeaderMonitorServer) error
	RaftState(context.Context, *serverpb.RaftStateRequest) (*serverpb.RaftStateResponse, error)
}

func (s *Server) responseHeader() *serverpb.ResponseHeader {
	return &serverpb.ResponseHeader{
		ClusterId: s.clusterID,
		RaftTerm:  atomic.LoadUint64(&s.term),
	}
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
		Namespace: req.Namespace,
		Key:       req.Key,
		Value: func() []byte {
			if req.IgnoreValue {
				return nil
			}
			return req.Value
		}(),
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &serverpb.SetResponse{Header: s.responseHeader()}, nil
}

func (s *Server) Get(_ context.Context, req *serverpb.GetRequest) (*serverpb.GetResponse, error) {
	storeAPI, err := s.currentNamespace(req.Namespace)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	value, err := storeAPI.Get(req.Key)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &serverpb.GetResponse{
		Header: s.responseHeader(),
		Value:  value.Data,
		Ttl:    nil, // unimplemented
	}, nil
}

func (s *Server) TrySet(ctx context.Context, req *serverpb.SetRequest) (*serverpb.SetResponse, error) {
	if _, err := s.raftApply(ctx, time.Millisecond*500, &LogPayload{
		Command: func() Command {
			if req.IgnoreTtl {
				return TrySet
			}
			return TrySetWithTTL
		}(),
		TTL: func() *uint32 {
			return &req.Ttl
		}(),
		Namespace: req.Namespace,
		Key:       req.Key,
		Value: func() []byte {
			if req.IgnoreValue {
				return nil
			}
			return req.Value
		}(),
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &serverpb.SetResponse{Header: s.responseHeader()}, nil
}

func (s *Server) Delete(ctx context.Context, req *serverpb.DeleteRequest) (*serverpb.DeleteResponse, error) {
	if _, err := s.raftApply(ctx, time.Millisecond*500, &LogPayload{
		Command:   Del,
		Namespace: req.Namespace,
		Key:       req.Key,
	}); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &serverpb.DeleteResponse{Header: s.responseHeader()}, nil
}

func (s *Server) Watch(req *serverpb.WatchRequest, stream serverpb.KV_WatchServer) error {
	storeAPI, err := s.currentNamespace(req.Namespace)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if !req.IgnoreErrors {
		if _, err = storeAPI.Get(req.Key); err == store.ErrKeyNotFound {
			return status.Error(codes.Aborted, err.Error())
		}
	}

	watcher, err := storeAPI.Watch(req.Key)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	defer watcher.Close()

	for {
		select {
		case value := <-watcher.Notify():
			if value.Deleted() && !req.IgnoreErrors {
				return status.Error(codes.Unavailable, "key has deleted")
			}
			if err = stream.Send(&serverpb.WatchResponse{
				Header:    s.responseHeader(),
				UpdateSeq: value.Seq,
				Timestamp: value.Timestamp,
				Data:      *value.Data,
			}); err != nil {
				// unrecoverable error
				return status.Error(codes.FailedPrecondition, err.Error())
			}
		}
	}

}

func (s *Server) Lock(_ context.Context, req *serverpb.LockRequest) (*serverpb.LockResponse, error) {
	if err := locker.MutexLock(req.LockId, req.Ttl, s.lockerBackend); err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return &serverpb.LockResponse{Header: s.responseHeader()}, nil
}

func (s *Server) Unlock(_ context.Context, req *serverpb.UnlockRequest) (*serverpb.UnlockResponse, error) {
	if err := locker.MutexUnlock(req.LockId, s.lockerBackend); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &serverpb.UnlockResponse{Header: s.responseHeader()}, nil
}

func (s *Server) TryLock(_ context.Context, req *serverpb.TryLockRequest) (*serverpb.TryLockResponse, error) {
	if !locker.MutexTryLock(req.LockId, req.Ttl, req.Deadline, s.lockerBackend) {
		return nil, status.Error(codes.PermissionDenied, locker.ErrStatusBusy.Error())
	}
	return &serverpb.TryLockResponse{Header: s.responseHeader()}, nil
}

func (s *Server) AppendCluster(_ context.Context, req *serverpb.AppendClusterRequest) (*serverpb.AppendClusterResponse, error) {
	if err := s.raft.AddCluster(raft.ServerID(req.ServerId), raft.ServerAddress(req.PeerAddr)); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return &serverpb.AppendClusterResponse{}, nil
}

func (s *Server) LeaderMonitor(_ *serverpb.LeaderMonitorRequest, stream serverpb.RedQueen_LeaderMonitorServer) error {
	var (
		notifyID = uuid.New().String()
		notify   = make(chan bool, 1)
	)
	s.stateNotify.Store(notifyID, notify)

	defer func() {
		s.stateNotify.Delete(notifyID)
		close(notify)
	}()
	s.raft.Stats()
	for {
		if err := stream.Send(&serverpb.LeaderMonitorResponse{
			Leader: <-notify,
		}); err != nil {
			return err
		}
	}
}

func (s *Server) RaftState(_ context.Context, _ *serverpb.RaftStateRequest) (*serverpb.RaftStateResponse, error) {
	var state serverpb.RaftState
	switch s.raft.Stats()["state"] {
	case "Follower":
		state = serverpb.RaftState_follower
	case "Candidate":
		state = serverpb.RaftState_candidate
	case "Leader":
		state = serverpb.RaftState_leader
	case "Shutdown":
		state = serverpb.RaftState_shutdown
	case "Unknown":
		state = serverpb.RaftState_unknown
	}
	return &serverpb.RaftStateResponse{State: state}, nil
}

package red

import (
	"context"
	"github.com/google/uuid"
	"github.com/hashicorp/raft"
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
	PrefixScan(context.Context, *serverpb.PrefixScanRequest) (*serverpb.PrefixScanResponse, error)
	TrySet(context.Context, *serverpb.SetRequest) (*serverpb.SetResponse, error)
	Delete(context.Context, *serverpb.DeleteRequest) (*serverpb.DeleteResponse, error)
	Watch(*serverpb.WatchRequest, serverpb.KV_WatchServer) error
	WatchPrefix(*serverpb.WatchPrefixRequest, serverpb.KV_WatchPrefixServer) error
}

type Locker interface {
	Lock(context.Context, *serverpb.LockRequest) (*serverpb.LockResponse, error)
	Unlock(context.Context, *serverpb.UnlockRequest) (*serverpb.UnlockResponse, error)
	TryLock(context.Context, *serverpb.TryLockRequest) (*serverpb.TryLockResponse, error)
}

type Internal interface {
	AppendCluster(context.Context, *serverpb.AppendClusterRequest) (*serverpb.AppendClusterResponse, error)
	LeaderMonitor(*serverpb.LeaderMonitorRequest, serverpb.RedQueen_LeaderMonitorServer) error
	RaftState(context.Context, *serverpb.RaftStateRequest) (*serverpb.RaftStateResponse, error)
}

func (s *Server) responseHeader() *serverpb.ResponseHeader {
	return &serverpb.ResponseHeader{
		ClusterId: s.clusterID,
		RaftTerm:  s.raft.Term(),
	}
}

func (s *Server) Set(ctx context.Context, req *serverpb.SetRequest) (*serverpb.SetResponse, error) {
	if err := s.applyLog(ctx, &serverpb.RaftLogPayload{
		Command: func() serverpb.RaftLogCommand {
			if req.IgnoreTtl {
				return serverpb.RaftLogCommand_Set
			}
			return serverpb.RaftLogCommand_SetWithTTL
		}(),
		Key: req.Key,
		Value: func() []byte {
			if req.IgnoreValue {
				return nil
			}
			return req.Value
		}(),
		Ttl: func() *uint32 {
			return &req.Ttl
		}(),
		Namespace: req.Namespace,
	}, time.Millisecond*500); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &serverpb.SetResponse{Header: s.responseHeader()}, nil
}

func (s *Server) Get(_ context.Context, req *serverpb.GetRequest) (*serverpb.GetResponse, error) {
	storeAPI, err := s.namespace(req.Namespace)
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
		Ttl:    value.TTL,
	}, nil
}

func (s *Server) PrefixScan(_ context.Context, req *serverpb.PrefixScanRequest) (*serverpb.PrefixScanResponse, error) {
	storeAPI, err := s.namespace(req.Namespace)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	scanResults, err := storeAPI.PrefixSearchScan(req.Prefix, req.GetReg(), int(req.Offset), int(req.Limit))
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &serverpb.PrefixScanResponse{
		Header: s.responseHeader(),
		Result: func() []*serverpb.PrefixScanResponse_PrefixScanResult {
			results := make([]*serverpb.PrefixScanResponse_PrefixScanResult, len(scanResults))
			for i, value := range scanResults {
				// todo: this part code will be removed future golang version
				// ---- start ----
				value := value
				// ---- end ----
				results[i] = &serverpb.PrefixScanResponse_PrefixScanResult{
					Key:       value.Key,
					Value:     value.Data,
					Timestamp: value.Timestamp,
					Ttl:       value.TTL,
				}
			}
			return results
		}(),
	}, nil
}

func (s *Server) TrySet(ctx context.Context, req *serverpb.SetRequest) (*serverpb.SetResponse, error) {
	if err := s.applyLog(ctx, &serverpb.RaftLogPayload{
		Command: func() serverpb.RaftLogCommand {
			if req.IgnoreTtl {
				return serverpb.RaftLogCommand_TrySet
			}
			return serverpb.RaftLogCommand_TrySetWithTTL
		}(),
		Key: req.Key,
		Value: func() []byte {
			if req.IgnoreValue {
				return nil
			}
			return req.Value
		}(),
		Ttl:       &req.Ttl,
		Namespace: req.Namespace,
	}, time.Millisecond*500); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &serverpb.SetResponse{Header: s.responseHeader()}, nil
}

func (s *Server) Delete(ctx context.Context, req *serverpb.DeleteRequest) (*serverpb.DeleteResponse, error) {
	if err := s.applyLog(ctx, &serverpb.RaftLogPayload{
		Command:   serverpb.RaftLogCommand_Del,
		Key:       req.Key,
		Namespace: req.Namespace,
	}, time.Millisecond*500); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &serverpb.DeleteResponse{Header: s.responseHeader()}, nil
}

func (s *Server) Watch(req *serverpb.WatchRequest, stream serverpb.KV_WatchServer) error {
	storeAPI, err := s.namespace(req.Namespace)
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
		value := <-watcher.Notify()
		if value.Deleted() && !req.IgnoreErrors {
			return status.Error(codes.Unavailable, "key has deleted")
		}
		if err = stream.Send(&serverpb.WatchResponse{
			Header:    s.responseHeader(),
			UpdateSeq: value.Seq,
			Timestamp: value.Timestamp,
			Ttl:       value.TTL,
			Key:       value.Key,
			Value: func() []byte {
				if value.Value == nil {
					return nil
				}
				return *value.Value
			}(),
		}); err != nil {
			// unrecoverable error
			return status.Error(codes.FailedPrecondition, err.Error())
		}
	}

}

func (s *Server) WatchPrefix(req *serverpb.WatchPrefixRequest, stream serverpb.KV_WatchPrefixServer) error {
	storeAPI, err := s.namespace(req.Namespace)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	watcher := storeAPI.WatchPrefix(req.Prefix)
	defer watcher.Close()

	for {
		value := <-watcher.Notify()
		if err = stream.Send(&serverpb.WatchResponse{
			Header:    s.responseHeader(),
			UpdateSeq: value.Seq,
			Timestamp: value.Timestamp,
			Ttl:       value.TTL,
			Key:       value.Key,
			Value: func() []byte {
				if value.Value == nil {
					return nil
				}
				return *value.Value
			}(),
		}); err != nil {
			// unrecoverable error
			return status.Error(codes.FailedPrecondition, err.Error())
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

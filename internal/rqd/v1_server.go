package rqd

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/RealFax/RedQueen/internal/rqd/store"
	"github.com/RealFax/RedQueen/pkg/dlocker"
	"github.com/RealFax/RedQueen/pkg/expr"
	"github.com/RealFax/RedQueen/pkg/fs"
	"github.com/RealFax/RedQueen/pkg/httputil"
	"github.com/google/uuid"
	"github.com/hashicorp/raft"
	"github.com/julienschmidt/httprouter"
	"google.golang.org/protobuf/types/known/emptypb"
	"io"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/RealFax/RedQueen/api/serverpb"
)

// ---- grpc handler ----

type v1RPCServer struct {
	*Server

	serverpb.UnimplementedKVServer
	serverpb.UnimplementedLockerServer
	serverpb.UnimplementedRedQueenServer
}

func (s *v1RPCServer) responseHeader() *serverpb.ResponseHeader {
	return &serverpb.ResponseHeader{
		ClusterId: s.clusterID,
		RaftTerm:  s.raft.Term(),
	}
}

func (s *v1RPCServer) Set(ctx context.Context, req *serverpb.SetRequest) (*serverpb.SetResponse, error) {
	if err := s.applyLog(ctx, &serverpb.RaftLogPayload{
		Command:   expr.If(req.IgnoreTtl, serverpb.RaftLogCommand_Set, serverpb.RaftLogCommand_SetWithTTL),
		Key:       req.Key,
		Value:     expr.If(req.IgnoreValue, nil, req.Value),
		Ttl:       expr.Pointer(req.Ttl),
		Namespace: req.Namespace,
	}, 500*time.Millisecond); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &serverpb.SetResponse{Header: s.responseHeader()}, nil
}

func (s *v1RPCServer) Get(_ context.Context, req *serverpb.GetRequest) (*serverpb.GetResponse, error) {
	act, err := s.trySwapContext(req.Namespace)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	value, err := act.Get(req.Key)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &serverpb.GetResponse{
		Header: s.responseHeader(),
		Value:  value.Data,
		Ttl:    value.TTL,
	}, nil
}

func (s *v1RPCServer) PrefixScan(_ context.Context, req *serverpb.PrefixScanRequest) (*serverpb.PrefixScanResponse, error) {
	act, err := s.trySwapContext(req.Namespace)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	scanResults, err := act.PrefixSearchScan(req.Prefix, req.GetReg(), int(req.Offset), int(req.Limit))
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

func (s *v1RPCServer) TrySet(ctx context.Context, req *serverpb.SetRequest) (*serverpb.SetResponse, error) {
	if err := s.applyLog(ctx, &serverpb.RaftLogPayload{
		Command:   expr.If(req.IgnoreTtl, serverpb.RaftLogCommand_TrySet, serverpb.RaftLogCommand_TrySetWithTTL),
		Key:       req.Key,
		Value:     expr.If(req.IgnoreValue, nil, req.Value),
		Ttl:       &req.Ttl,
		Namespace: req.Namespace,
	}, 500*time.Millisecond); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &serverpb.SetResponse{Header: s.responseHeader()}, nil
}

func (s *v1RPCServer) Delete(ctx context.Context, req *serverpb.DeleteRequest) (*serverpb.DeleteResponse, error) {
	if err := s.applyLog(ctx, &serverpb.RaftLogPayload{
		Command:   serverpb.RaftLogCommand_Del,
		Key:       req.Key,
		Namespace: req.Namespace,
	}, 500*time.Millisecond); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &serverpb.DeleteResponse{Header: s.responseHeader()}, nil
}

func (s *v1RPCServer) Watch(req *serverpb.WatchRequest, stream serverpb.KV_WatchServer) error {
	storeAPI, err := s.trySwapContext(req.Namespace)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	if !req.IgnoreErrors {
		if _, err = storeAPI.Get(req.Key); errors.Is(err, store.ErrKeyNotFound) {
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

func (s *v1RPCServer) WatchPrefix(req *serverpb.WatchPrefixRequest, stream serverpb.KV_WatchPrefixServer) error {
	storeAPI, err := s.trySwapContext(req.Namespace)
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

func (s *v1RPCServer) Lock(_ context.Context, req *serverpb.LockRequest) (*serverpb.LockResponse, error) {
	if err := dlocker.MutexLock(req.LockId, req.Ttl, s.lockerBackend); err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return &serverpb.LockResponse{Header: s.responseHeader()}, nil
}

func (s *v1RPCServer) Unlock(_ context.Context, req *serverpb.UnlockRequest) (*serverpb.UnlockResponse, error) {
	if err := dlocker.MutexUnlock(req.LockId, s.lockerBackend); err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &serverpb.UnlockResponse{Header: s.responseHeader()}, nil
}

func (s *v1RPCServer) TryLock(_ context.Context, req *serverpb.TryLockRequest) (*serverpb.TryLockResponse, error) {
	if !dlocker.MutexTryLock(req.LockId, req.Ttl, req.Deadline, s.lockerBackend) {
		return nil, status.Error(codes.PermissionDenied, dlocker.ErrStatusBusy.Error())
	}
	return &serverpb.TryLockResponse{Header: s.responseHeader()}, nil
}

func (s *v1RPCServer) AppendCluster(_ context.Context, req *serverpb.AppendClusterRequest) (*serverpb.AppendClusterResponse, error) {
	if err := s.raft.AddCluster(raft.ServerID(req.ServerId), raft.ServerAddress(req.PeerAddr)); err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return &serverpb.AppendClusterResponse{}, nil
}

func (s *v1RPCServer) LeaderMonitor(_ *serverpb.LeaderMonitorRequest, stream serverpb.RedQueen_LeaderMonitorServer) error {
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

func (s *v1RPCServer) RaftState(_ context.Context, _ *emptypb.Empty) (*serverpb.RaftStateResponse, error) {
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

func (s *v1RPCServer) RaftSnapshot(_ context.Context, req *serverpb.RaftSnapshotRequest) (*emptypb.Empty, error) {
	future := s.raft.Snapshot()
	if future.Error() != nil {
		return nil, status.Error(codes.Internal, future.Error().Error())
	}

	if req.Path == nil {
		return &emptypb.Empty{}, nil
	}

	_, rc, err := future.Open()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rc.Close()

	f, err := fs.MustOpen(req.GetPath())
	if err != nil {
		return nil, status.Errorf(codes.Aborted, "server open snapshot file error, cause: %s", err)
	}
	defer f.Close()

	if _, err = io.Copy(f, rc); err != nil {
		return nil, status.Errorf(codes.Aborted, "server write snapshot file error, cause: %s", err)
	}

	return &emptypb.Empty{}, nil
}

// ---- http handler ----

type v1HttpServer struct {
	*Server
}

func (s *v1HttpServer) getBucket(c context.Context) *string {
	params, ok := c.Value(httprouter.ParamsKey).(httprouter.Params)
	if !ok {
		return nil
	}
	if bucket := params.ByName("bucket"); bucket != "" {
		return &bucket
	}
	return nil
}

func (s *v1HttpServer) responseHeader(w http.ResponseWriter) {
	w.Header().Add("X-Cluster-ID", s.clusterID)
	w.Header().Set("X-Raft-Term", strconv.FormatUint(s.raft.Term(), 10))
}

func (s *v1HttpServer) Set(w http.ResponseWriter, r *http.Request) error {
	req, err := httputil.XBindJSON[*serverpb.SetRequest](r.Body)
	if err != nil {
		return err
	}

	if err = s.applyLog(r.Context(), &serverpb.RaftLogPayload{
		Command:   expr.If(req.IgnoreTtl, serverpb.RaftLogCommand_Set, serverpb.RaftLogCommand_SetWithTTL),
		Key:       req.Key,
		Value:     expr.If(req.IgnoreValue, nil, req.Value),
		Ttl:       expr.Pointer(req.Ttl),
		Namespace: s.getBucket(r.Context()),
	}, 500*time.Millisecond); err != nil {
		return httputil.StatusWrap(http.StatusInternalServerError, 0, err)
	}

	defer s.responseHeader(w)
	httputil.Any(http.StatusCreated, 1).Ok(w)
	return nil
}

func (s *v1HttpServer) Get(w http.ResponseWriter, r *http.Request) error {
	_key := r.URL.Query().Get("key")
	if _key == "" {
		return httputil.NewStatus(http.StatusBadRequest, 0, "invalid query key")
	}
	key, err := base64.URLEncoding.DecodeString(_key)
	if err != nil {
		return httputil.NewStatus(http.StatusBadRequest, 0, "invalid query key")
	}

	act, err := s.trySwapContext(s.getBucket(r.Context()))
	if err != nil {
		return httputil.StatusWrap(http.StatusInternalServerError, 0, err)
	}

	value, err := act.Get(key)
	if err != nil {
		return httputil.StatusWrap(http.StatusNotFound, -1, err)
	}

	defer s.responseHeader(w)
	httputil.NewAck[*serverpb.GetResponse](http.StatusOK, 1).Data(&serverpb.GetResponse{
		Value: value.Data,
		Ttl:   value.TTL,
	}).Ok(w)
	return nil
}

func (s *v1HttpServer) PrefixScan(w http.ResponseWriter, r *http.Request) error {
	var (
		q      = r.URL.Query()
		prefix []byte
		offset uint64
		limit  uint64
	)
	prefix, _ = base64.URLEncoding.DecodeString(q.Get("prefix"))
	offset, _ = strconv.ParseUint(q.Get("offset"), 10, 64)
	limit, _ = strconv.ParseUint(q.Get("limit"), 10, 74)

	if limit == 0 {
		// limit is zero, should init
		limit = 10
	}

	act, err := s.trySwapContext(s.getBucket(r.Context()))
	if err != nil {
		return httputil.StatusWrap(http.StatusInternalServerError, 0, err)
	}

	scanResults, err := act.PrefixSearchScan(prefix, q.Get("reg"), int(offset), int(limit))
	if err != nil {
		return httputil.StatusWrap(http.StatusNotFound, -1, err)
	}

	defer s.responseHeader(w)
	httputil.NewAck[[]*serverpb.PrefixScanResponse_PrefixScanResult](http.StatusOK, 1).
		Data(func() []*serverpb.PrefixScanResponse_PrefixScanResult {
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
		}()).Ok(w)
	return nil
}

func (s *v1HttpServer) TrySet(w http.ResponseWriter, r *http.Request) error {
	req, err := httputil.XBindJSON[*serverpb.SetRequest](r.Body)
	if err != nil {
		return err
	}

	if err = s.applyLog(r.Context(), &serverpb.RaftLogPayload{
		Command:   expr.If(req.IgnoreTtl, serverpb.RaftLogCommand_TrySet, serverpb.RaftLogCommand_TrySetWithTTL),
		Key:       req.Key,
		Value:     expr.If(req.IgnoreValue, nil, req.Value),
		Ttl:       &req.Ttl,
		Namespace: s.getBucket(r.Context()),
	}, 500*time.Millisecond); err != nil {
		return httputil.StatusWrap(http.StatusInternalServerError, 0, err)
	}

	defer s.responseHeader(w)
	httputil.Any(http.StatusCreated, 1).Ok(w)
	return nil
}

func (s *v1HttpServer) Delete(w http.ResponseWriter, r *http.Request) error {
	req, err := httputil.XBindJSON[*serverpb.DeleteRequest](r.Body)
	if err != nil {
		return err
	}

	if err = s.applyLog(r.Context(), &serverpb.RaftLogPayload{
		Command:   serverpb.RaftLogCommand_Del,
		Key:       req.Key,
		Namespace: s.getBucket(r.Context()),
	}, 500*time.Millisecond); err != nil {
		return httputil.StatusWrap(http.StatusInternalServerError, 0, err)
	}

	defer s.responseHeader(w)
	httputil.Any(http.StatusOK, 1).Ok(w)
	return nil
}

func (s *v1HttpServer) Lock(w http.ResponseWriter, r *http.Request) error {
	req, err := httputil.XBindJSON[*serverpb.LockRequest](r.Body)
	if err != nil {
		return err
	}

	if err = dlocker.MutexLock(req.LockId, req.Ttl, s.lockerBackend); err != nil {
		return httputil.StatusWrap(http.StatusForbidden, 0, err)
	}

	defer s.responseHeader(w)
	httputil.Any(http.StatusOK, 1).Ok(w)
	return nil
}

func (s *v1HttpServer) Unlock(w http.ResponseWriter, r *http.Request) error {
	req, err := httputil.XBindJSON[*serverpb.UnlockRequest](r.Body)
	if err != nil {
		return err
	}

	if err = dlocker.MutexUnlock(req.LockId, s.lockerBackend); err != nil {
		return httputil.StatusWrap(http.StatusNotFound, 0, err)
	}

	defer s.responseHeader(w)
	httputil.Any(http.StatusOK, 1).Ok(w)
	return nil
}

func (s *v1HttpServer) TryLock(w http.ResponseWriter, r *http.Request) error {
	req, err := httputil.XBindJSON[*serverpb.TryLockRequest](r.Body)
	if err != nil {
		return err
	}

	if !dlocker.MutexTryLock(req.LockId, req.Ttl, req.Deadline, s.lockerBackend) {
		return httputil.StatusWrap(http.StatusForbidden, 0, dlocker.ErrStatusBusy)
	}

	defer s.responseHeader(w)
	httputil.Any(http.StatusOK, 1).Ok(w)
	return nil
}

func (s *v1HttpServer) AppendCluster(w http.ResponseWriter, r *http.Request) error {
	req, err := httputil.XBindJSON[*serverpb.AppendClusterRequest](r.Body)
	if err != nil {
		return err
	}

	if err = s.raft.AddCluster(raft.ServerID(req.ServerId), raft.ServerAddress(req.PeerAddr)); err != nil {
		return httputil.StatusWrap(http.StatusForbidden, 0, err)
	}

	defer s.responseHeader(w)
	httputil.Any(http.StatusCreated, 1).Ok(w)
	return nil
}

func (s *v1HttpServer) Stats(w http.ResponseWriter, _ *http.Request) error {
	s.responseHeader(w)
	h := s.raft.Stats()
	leaderAddr, leaderID := s.raft.LeaderWithID()
	h["leader_address"], h["leader_id"] = string(leaderAddr), string(leaderID)
	httputil.NewAck[map[string]string](http.StatusOK, 1).Data(h).Ok(w)
	return nil
}

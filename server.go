package RedQueen

import (
	"context"
	"github.com/RealFax/RedQueen/locker"
	"github.com/RealFax/RedQueen/store"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
	"time"
)

type Server struct {
	term      uint64 // [ATOMIC]
	clusterID uint64

	store         store.Store
	lockerBackend locker.Backend

	raft *Raft
}

func (s *Server) currentNamespace(namespace *string) (store.Namespace, error) {
	var (
		err      error
		storeAPI store.Namespace = s.store
	)

	if namespace != nil {
		if storeAPI, err = s.store.Namespace(*namespace); err != nil {
			return nil, err
		}
	}

	return storeAPI, nil
}

func (s *Server) raftApply(ctx context.Context, timeout time.Duration, lp *LogPayload) (raft.ApplyFuture, error) {
	cmd, err := msgpack.Marshal(lp)
	if err != nil {
		return nil, err
	}

	var (
		ch = make(chan raft.ApplyFuture, 1)
	)

	go func() {
		ch <- s.raft.Apply(cmd, timeout)
		close(ch)
	}()

	select {
	case <-ctx.Done():
		return nil, errors.Wrap(ctx.Err(), "context canceled")
	case x := <-ch:
		err = x.Error()
		return x, err
	}
}

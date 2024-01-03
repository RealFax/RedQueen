package red

import (
	"bytes"
	"io"
	"sync/atomic"

	"github.com/hashicorp/raft"
	"github.com/pkg/errors"

	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/store"
)

type FSMHandleFunc func(*serverpb.RaftLogPayload) error
type FSM struct {
	Term     *uint64
	Handlers map[serverpb.RaftLogCommand]FSMHandleFunc
	Store    store.Store
}

func (f *FSM) Apply(log *raft.Log) any {
	atomic.StoreUint64(f.Term, log.Term)
	switch log.Type {
	case raft.LogCommand:
		messages, err := UnpackLog(bytes.NewReader(log.Data))
		if err != nil {
			return err
		}
		for _, message := range messages {
			handle, ok := f.Handlers[message.Command]
			if !ok {
				return errors.Errorf("unimplemented command %s handler", message.Command.String())
			}
			if err = handle(message); err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	snapshot, err := f.Store.Snapshot()
	if err != nil {
		return nil, errors.Wrap(err, "create snapshot fail")
	}
	return &Snapshot{snapshot}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	return f.Store.Restore(rc)
}

package red

import (
	"bytes"
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/store"

	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"io"
)

type FSMHandleFunc func(*serverpb.RaftLogPayload) error
type FSM struct {
	handlers map[serverpb.RaftLogCommand]FSMHandleFunc
	store    store.Store
}

func (f *FSM) Apply(log *raft.Log) any {
	switch log.Type {
	case raft.LogCommand:
		msgs, err := UnpackLog(bytes.NewReader(log.Data))
		if err != nil {
			return err
		}
		for _, msg := range msgs {
			handle, ok := f.handlers[msg.Command]
			if !ok {
				return errors.New("there's no corresponding command handle")
			}
			if err = handle(msg); err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	snapshot, err := f.store.Snapshot()
	if err != nil {
		return nil, errors.Wrap(err, "create snapshot fail")
	}
	return &Snapshot{snapshot}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	return f.store.Restore(rc)
}

func NewFSM(s store.Store, handlers map[serverpb.RaftLogCommand]FSMHandleFunc) *FSM {
	return &FSM{
		handlers: handlers,
		store:    s,
	}
}

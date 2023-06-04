package red

import (
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/store"

	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"io"
)

type FSMHandleFunc func(*serverpb.RaftLogPayload) error

type FSM struct {
	handles map[serverpb.RaftLogCommand]FSMHandleFunc
	store   store.Store
}

func (f *FSM) namespaceSwitch(namespace *string) (store.Namespace, error) {
	if namespace == nil {
		return f.store, nil
	}
	storeAPI, err := f.store.Namespace(*namespace)
	if err != nil {
		return nil, errors.Wrap(err, "namespace switch error")
	}
	return storeAPI, nil
}

func (f *FSM) hSetWithTTL(payload *serverpb.RaftLogPayload) error {
	if payload.Ttl == nil || payload.Key == nil || payload.Value == nil {
		return errors.New("invalid SetWithTTL args")
	}

	dest, err := f.namespaceSwitch(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.SetWithTTL(payload.Key, payload.Value, *payload.Ttl)
}

func (f *FSM) hTrySetWithTTL(payload *serverpb.RaftLogPayload) error {
	if payload.Ttl == nil || payload.Key == nil || payload.Value == nil {
		return errors.New("invalid TrySetWithTTL args")
	}

	dest, err := f.namespaceSwitch(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.TrySetWithTTL(payload.Key, payload.Value, *payload.Ttl)
}

func (f *FSM) hSet(payload *serverpb.RaftLogPayload) error {
	if payload.Key == nil || payload.Value == nil {
		return errors.New("invalid Set args")
	}

	dest, err := f.namespaceSwitch(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.Set(payload.Key, payload.Value)
}

func (f *FSM) hTrySet(payload *serverpb.RaftLogPayload) error {
	if payload.Key == nil || payload.Value == nil {
		return errors.New("invalid TrySet args")
	}

	dest, err := f.namespaceSwitch(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.TrySet(payload.Key, payload.Value)
}

func (f *FSM) hDel(payload *serverpb.RaftLogPayload) error {
	if payload.Key == nil {
		return errors.New("invalid Del args")
	}

	dest, err := f.namespaceSwitch(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.Del(payload.Key)
}

func (f *FSM) Apply(log *raft.Log) any {
	switch log.Type {
	case raft.LogCommand:
		var msg serverpb.RaftLogPayload
		if err := proto.Unmarshal(log.Data, &msg); err != nil {
			return errors.Wrap(err, "unmarshal proto error:")
		}
		handle, ok := f.handles[msg.Command]
		if !ok {
			return errors.New("there's no corresponding command handle")
		}
		return handle(&msg)
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
	return rc.Close()
}

func NewFSM(s store.Store) *FSM {
	fsm := &FSM{store: s}

	fsm.handles = map[serverpb.RaftLogCommand]FSMHandleFunc{
		serverpb.RaftLogCommand_SetWithTTL:    fsm.hSetWithTTL,
		serverpb.RaftLogCommand_TrySetWithTTL: fsm.hTrySetWithTTL,
		serverpb.RaftLogCommand_Set:           fsm.hSet,
		serverpb.RaftLogCommand_TrySet:        fsm.hTrySet,
		serverpb.RaftLogCommand_Del:           fsm.hDel,
	}

	return fsm
}

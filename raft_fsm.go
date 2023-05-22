package RedQueen

import (
	"github.com/RealFax/RedQueen/store"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
	"io"
)

type FSM struct {
	handler map[Command]func(payload *LogPayload) error
	store   store.Store
}

func (f *FSM) currentNamespace(namespace *string) (store.Namespace, error) {
	var (
		err      error
		storeAPI store.Namespace = f.store
	)

	if namespace != nil {
		if storeAPI, err = f.store.Namespace(*namespace); err != nil {
			return nil, errors.Wrap(err, "switch namespace fatal")
		}
	}

	return storeAPI, nil
}

func (f *FSM) handleSetWithTTL(payload *LogPayload) error {
	if payload.TTL == nil || payload.Key == nil || payload.Value == nil {
		return errors.New("invalid SetWithTTL args")
	}

	dest, err := f.currentNamespace(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.SetWithTTL(payload.Key, payload.Value, *payload.TTL)
}

func (f *FSM) handleTrySetWithTTL(payload *LogPayload) error {
	if payload.TTL == nil || payload.Key == nil || payload.Value == nil {
		return errors.New("invalid TrySetWithTTL args")
	}

	dest, err := f.currentNamespace(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.TrySetWithTTL(payload.Key, payload.Value, *payload.TTL)
}

func (f *FSM) handleSet(payload *LogPayload) error {
	if payload.Key == nil || payload.Value == nil {
		return errors.New("invalid Set args")
	}

	dest, err := f.currentNamespace(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.Set(payload.Key, payload.Value)
}

func (f *FSM) handleTrySet(payload *LogPayload) error {
	if payload.Key == nil || payload.Value == nil {
		return errors.New("invalid TrySet args")
	}

	dest, err := f.currentNamespace(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.TrySet(payload.Key, payload.Value)
}

func (f *FSM) handleDel(payload *LogPayload) error {
	if payload.Key == nil {
		return errors.New("invalid Del args")
	}

	dest, err := f.currentNamespace(payload.Namespace)
	if err != nil {
		return err
	}

	return dest.Del(payload.Key)
}

func (f *FSM) Apply(log *raft.Log) any {
	switch log.Type {
	case raft.LogCommand:
		var payload LogPayload
		if err := msgpack.Unmarshal(log.Data, &payload); err != nil {
			return errors.Wrap(err, "could not parse payload")
		}
		handle, ok := f.handler[payload.Command]
		if !ok {
			return errors.New("there's no corresponding command handle")
		}
		return handle(&payload)
	default:
		return errors.Errorf("unknonw raft log type: %s", log.Type.String())
	}
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	snapshot, err := f.store.Snapshot()
	if err != nil {
		return nil, errors.Wrap(err, "create snapshot failed")
	}
	return &Snapshot{snapshot}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {

	return rc.Close()
}

func NewFSM(s store.Store) *FSM {
	fsm := &FSM{store: s}

	// register fsm handlers
	fsm.handler = map[Command]func(payload *LogPayload) error{
		SetWithTTL:    fsm.handleSetWithTTL,
		TrySetWithTTL: fsm.handleTrySetWithTTL,
		Set:           fsm.handleSet,
		TrySet:        fsm.handleTrySet,
		Del:           fsm.handleDel,
	}

	return fsm
}

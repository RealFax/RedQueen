package RedQueen

import (
	"encoding/json"
	"github.com/RealFax/RedQueen/store"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"io"
)

type FSM struct {
	Store store.Store
}

func (f *FSM) Apply(log *raft.Log) any {
	switch log.Type {
	case raft.LogCommand:
		var payload LogPayload
		if err := json.Unmarshal(log.Data, &payload); err != nil {
			return errors.Wrap(err, "could not parse payload")
		}
		return f.Store.Set(payload.Key, payload.Value)
	default:
		return errors.Errorf("unknonw raft log type: %s", log.Type.String())
	}
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return &Snapshot{}, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	dec := json.NewDecoder(rc)
	if dec.More() {
		var payload LogPayload
		if err := dec.Decode(&payload); err != nil {
			return errors.Wrap(err, "could not decode payload")
		}
		f.Store.Set(payload.Key, payload.Value)
	}
	return rc.Close()
}

func NewFSM(s store.Store) *FSM {
	return &FSM{Store: s}
}

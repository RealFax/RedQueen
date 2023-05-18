package RedQueen

import (
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/pkg/errors"
	"net"
	"os"
	"time"
)

type Command int32

const (
	SetWithTTL Command = iota
	TrySetWithTTL
	Set
	TrySet
	Del
)

type LogPayload struct {
	Command   Command `json:"command" msgpack:"cmd"`
	TTL       *uint32 `json:"ttl,omitempty" msgpack:"ttl,omitempty"`
	Namespace *string `json:"namespace,omitempty" msgpack:"ns,omitempty"`
	Key       []byte  `json:"key,omitempty" msgpack:"k,omitempty"`
	Value     []byte  `json:"value,omitempty" msgpack:"v,omitempty"`
}

type RaftConfig struct {
	ServerID, Addr, BoltStorePath, FileSnapshotStorePath string
	FSM                                                  *FSM
	Clusters                                             []raft.Server
}

type Raft struct {
	*raft.Raft
}

func (r *Raft) AddCluster(id raft.ServerID, addr raft.ServerAddress) error {
	return r.AddVoter(id, addr, 0, 0).Error()
}

func NewRaft(bootstrap bool, rcfg RaftConfig) (*Raft, error) {
	cfg := raft.DefaultConfig()
	cfg.LocalID = raft.ServerID(rcfg.ServerID)

	store, err := raftboltdb.NewBoltStore(rcfg.BoltStorePath)
	if err != nil {
		return nil, errors.Wrap(err, "boltdb")
	}

	snapshot, err := raft.NewFileSnapshotStore(rcfg.FileSnapshotStorePath, 2, os.Stderr)
	if err != nil {
		return nil, errors.Wrap(err, "snapshot")
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", rcfg.Addr)
	if err != nil {
		return nil, errors.Wrap(err, "resolve tcp addr")
	}

	transport, err := raft.NewTCPTransport(rcfg.Addr, tcpAddr, 10, time.Second*5, os.Stderr)
	if err != nil {
		return nil, errors.Wrap(err, "tcp transport")
	}

	r, err := raft.NewRaft(cfg, rcfg.FSM, store, store, snapshot, transport)
	if err != nil {
		return nil, errors.Wrap(err, "raft")
	}

	if bootstrap {
		if err = r.BootstrapCluster(raft.Configuration{
			Servers: rcfg.Clusters,
		}).Error(); err != nil {
			return nil, err
		}
	}

	return &Raft{r}, nil
}

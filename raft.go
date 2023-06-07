package red

import (
	"github.com/RealFax/RedQueen/store"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/pkg/errors"
	"net"
	"os"
	"path/filepath"
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
	MaxSnapshots            int
	ServerID, Addr, DataDir string
	Store                   store.Store
	Clusters                []raft.Server
}

type Raft struct {
	*raft.Raft
}

func (r *Raft) AddCluster(id raft.ServerID, addr raft.ServerAddress) error {
	return r.AddVoter(id, addr, 0, 0).Error()
}

func NewRaft(bootstrap bool, cfg RaftConfig) (*Raft, error) {
	raftCfg := raft.DefaultConfig()
	raftCfg.LocalID = raft.ServerID(cfg.ServerID)

	raftCfg.LogLevel = "INFO"

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft-log.db"))
	if err != nil {
		return nil, errors.Wrap(err, "raft-log")
	}

	stableStore, err := NewStableStore(cfg.Store)
	if err != nil {
		return nil, errors.Wrap(err, "raft-stable-store")
	}

	snapshot, err := raft.NewFileSnapshotStore(cfg.DataDir, cfg.MaxSnapshots, os.Stderr)
	if err != nil {
		return nil, errors.Wrap(err, "snapshot")
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", cfg.Addr)
	if err != nil {
		return nil, errors.Wrap(err, "resolve-tcp-addr")
	}

	transport, err := raft.NewTCPTransport(cfg.Addr, tcpAddr, 10, time.Second*5, os.Stderr)
	if err != nil {
		return nil, errors.Wrap(err, "tcp-transport")
	}

	r, err := raft.NewRaft(raftCfg, NewFSM(cfg.Store), logStore, stableStore, snapshot, transport)
	if err != nil {
		return nil, errors.Wrap(err, "raft")
	}

	if bootstrap {
		if err = r.BootstrapCluster(raft.Configuration{
			Servers: cfg.Clusters,
		}).Error(); err != nil {
			return nil, err
		}
	}

	return &Raft{r}, nil
}

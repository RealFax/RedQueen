package RedQueen

import (
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/pkg/errors"
	"net"
	"os"
	"time"
)

type LogPayload struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

type RaftConfig struct {
	ServerID, Addr, BoldStorePath, FileSnapshotStorePath string
	FSM                                                  *FSM
	Clusters                                             []raft.Server
}

type Raft struct {
	Raft *raft.Raft
}

func (r *Raft) AppendCluster(id raft.ServerID, addr raft.ServerAddress) error {
	return r.Raft.AddVoter(id, addr, 0, 0).Error()
}

func NewRaft(rcfg RaftConfig) (*Raft, error) {
	cfg := raft.DefaultConfig()
	cfg.LocalID = raft.ServerID(rcfg.ServerID)

	store, err := raftboltdb.NewBoltStore(rcfg.BoldStorePath)
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

	return &Raft{Raft: r}, r.BootstrapCluster(raft.Configuration{
		Servers: rcfg.Clusters,
	}).Error()
}

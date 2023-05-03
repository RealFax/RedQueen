package RedQueen

import "github.com/hashicorp/raft"

type Snapshot struct{}

func (s *Snapshot) Persist(_ raft.SnapshotSink) error { return nil }
func (s *Snapshot) Release()                          {}

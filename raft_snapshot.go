package red

import (
	"github.com/hashicorp/raft"
	"io"
)

type Snapshot struct {
	io.Reader
}

func (s *Snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, s); err != nil {
		_ = sink.Cancel()
		return err
	}
	return sink.Close()
}

func (s *Snapshot) Release() {}

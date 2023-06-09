package red

import (
	"bytes"
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

func (s *Snapshot) Release() {
	p, ok := s.Reader.(*bytes.Buffer)
	if !ok {
		return
	}
	p.Reset()
	p = nil
}

package red_test

import (
	red "github.com/RealFax/RedQueen"
	"github.com/RealFax/RedQueen/api/serverpb"
	"testing"
)

var raftLogPayloadMessage = &serverpb.RaftLogPayload{
	Command: serverpb.RaftLogCommand_TrySet,
	Key:     []byte("test_key"),
}

func init() {
	n := "test_namespace"
	raftLogPayloadMessage.Namespace = &n
}

func TestRaftLogPayloadKey(t *testing.T) {
	red.RaftLogPayloadKey(raftLogPayloadMessage)
}

func BenchmarkRaftLogPayloadKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		red.RaftLogPayloadKey(raftLogPayloadMessage)
	}
}

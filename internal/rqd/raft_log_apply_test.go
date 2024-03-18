package rqd_test

import (
	"github.com/RealFax/RedQueen/api/serverpb"
	red "github.com/RealFax/RedQueen/internal/rqd"
	"github.com/RealFax/RedQueen/pkg/expr"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	raftLogPayloadMessage = &serverpb.RaftLogPayload{
		Command:   serverpb.RaftLogCommand_TrySet,
		Key:       []byte("test_key"),
		Namespace: expr.Pointer("test_namespace"),
	}
	expectPayloadKey = uint64(4855146586712729396)
)

func TestRaftLogPayloadKey(t *testing.T) {
	assert.Equal(t, expectPayloadKey, red.RaftLogPayloadKey(raftLogPayloadMessage))
}

func BenchmarkRaftLogPayloadKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		red.RaftLogPayloadKey(raftLogPayloadMessage)
	}
}

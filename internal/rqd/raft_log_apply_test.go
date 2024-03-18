package rqd_test

import (
	"context"
	"github.com/RealFax/RedQueen/api/serverpb"
	red "github.com/RealFax/RedQueen/internal/rqd"
	"github.com/RealFax/RedQueen/pkg/expr"
	"github.com/hashicorp/raft"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var _ raft.ApplyFuture = &future{}

type future struct {
	response any
	error    error
}

func (f future) Index() uint64 { return 0 }
func (f future) Response() any { return f.response }
func (f future) Error() error  { return f.error }

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

func TestRaftSingleLogApplyer_Apply(t *testing.T) {
	applyer := red.NewRaftSingeLogApply(func(cmd []byte, timeout time.Duration) raft.ApplyFuture {
		assert.Equal(t, 1*time.Second, timeout)
		return &future{}
	})
	assert.Equal(t, applyer.Apply(nil, &serverpb.RaftLogPayload{}, 1*time.Second), red.ErrApplyLogDone)
}

func TestRaftMultipleLogApply_Apply(t *testing.T) {
	raftApply := red.NewRaftMultipleLogApply(
		context.Background(),
		10,
		5*time.Second,
		2*time.Second,
		func(data []byte, timeout time.Duration) raft.ApplyFuture {
			return &future{}
		},
	)
	err := raftApply.Apply(expr.Pointer(context.Background()), &serverpb.RaftLogPayload{}, time.Second)
	assert.NoError(t, err)
}

func BenchmarkRaftLogPayloadKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		red.RaftLogPayloadKey(raftLogPayloadMessage)
	}
}

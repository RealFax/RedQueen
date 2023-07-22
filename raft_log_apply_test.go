package red_test

import (
	"bytes"
	"context"
	red "github.com/RealFax/RedQueen"
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/internal/collapsar"
	"github.com/hashicorp/raft"
	"google.golang.org/protobuf/proto"
	"io"
	"strconv"
	"testing"
	"time"
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
	t.Logf("Payload key: %d", red.RaftLogPayloadKey(raftLogPayloadMessage))
}

func TestRaftMultipleLogApply_Apply(t *testing.T) {
	rm := red.NewRaftMultipleLogApply(
		context.Background(),
		3,
		time.Millisecond*300,
		time.Millisecond*300,
		func(cmd []byte, timeout time.Duration) raft.ApplyFuture {
			r, err := collapsar.NewReader(bytes.NewReader(cmd[4:]))
			if err != nil {
				t.Fatal(err)
			}

			for {
				pack, rErr := r.Next()
				if rErr != nil {
					if rErr == io.EOF {
						break
					}
					t.Fatal(err)
				}
				m := &serverpb.RaftLogPayload{}
				if err = proto.Unmarshal(pack, m); err != nil {
					t.Fatal(err)
				}
				t.Log(m.String())
			}
			return nil
		})

	for i := 0; i < 3; i++ {
		ctx := context.Background()
		if err := rm.Apply(&ctx, raftLogPayloadMessage, 0); err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Second * 1)

	for i := 0; i < 3; i++ {
		ctx := context.Background()
		raftLogPayloadMessage.Key = []byte("test_key" + strconv.Itoa(i))
		if err := rm.Apply(&ctx, raftLogPayloadMessage, 0); err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Second * 1)

}

func BenchmarkRaftLogPayloadKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		red.RaftLogPayloadKey(raftLogPayloadMessage)
	}
}

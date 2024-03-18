package grpcutil_test

import (
	"context"
	"github.com/RealFax/RedQueen/pkg/grpcutil"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"testing"
)

type mockClientStream struct {
	grpc.ClientStream
	ctx context.Context
}

func (m *mockClientStream) Context() context.Context {
	return m.ctx
}

func TestBasicAuthClientUnary(t *testing.T) {
	username := "username"
	password := "password"

	client := grpcutil.NewBasicAuthClient(username, password)

	invoker := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, client.AuthKey, md.Get(grpcutil.MetadataAuthorization)[0])
		return nil
	}

	err := client.Unary(context.Background(), "Method", nil, nil, &grpc.ClientConn{}, invoker)

	assert.NoError(t, err)
}

func TestBasicAuthClientStream(t *testing.T) {
	username := "username"
	password := "password"

	client := grpcutil.NewBasicAuthClient(username, password)

	streamer := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, client.AuthKey, md.Get(grpcutil.MetadataAuthorization)[0])

		return &mockClientStream{ctx: ctx}, nil
	}

	stream, err := client.Stream(context.Background(), nil, &grpc.ClientConn{}, "Method", streamer)

	// 验证结果
	assert.NoError(t, err)
	assert.NotNil(t, stream)
}

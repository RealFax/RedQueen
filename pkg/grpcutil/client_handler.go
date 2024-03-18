package grpcutil

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type BasicAuthClient struct {
	AuthKey string
}

func (c BasicAuthClient) ctxWrap(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.MD{
		MetadataAuthorization: []string{c.AuthKey},
	})
}

func (c BasicAuthClient) Unary(
	ctx context.Context,
	method string, req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return invoker(c.ctxWrap(ctx), method, req, reply, cc, opts...)
}

func (c BasicAuthClient) Stream(
	ctx context.Context,
	desc *grpc.StreamDesc,
	cc *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	return streamer(c.ctxWrap(ctx), desc, cc, method, opts...)
}

func NewBasicAuthClient(username, password string) *BasicAuthClient {
	return &BasicAuthClient{AuthKey: BuildAuthorization(username, password)}
}

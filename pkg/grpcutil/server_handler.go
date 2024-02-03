package grpcutil

import (
	"context"
	"crypto/subtle"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	MetadataAuthorization string = "Authorization"
)

type BasicAuthFunc func(username, password string) bool
type BasicAuth struct {
	authFC BasicAuthFunc
}

func (a BasicAuth) auth(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.InvalidArgument, "failed get metadata")
	}

	authorization := md.Get(MetadataAuthorization)
	if len(authorization) != 1 {
		return status.Error(codes.InvalidArgument, "invalid metadata 'Authorization'")
	}

	if !ParseAuthorization(authorization[0], a.authFC) {
		return status.Error(codes.Unauthenticated, "unauthenticated")
	}

	return nil
}

func (a BasicAuth) Unary(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if err := a.auth(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (a BasicAuth) Stream(
	srv any,
	ss grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	if err := a.auth(ss.Context()); err != nil {
		return err
	}
	return handler(srv, ss)
}

func NewBasicAuth(fc BasicAuthFunc) *BasicAuth {
	return &BasicAuth{authFC: fc}
}

func NewMemoryBasicAuthFunc(users map[string]string) BasicAuthFunc {
	return func(username, password string) bool {
		return subtle.ConstantTimeCompare([]byte(users[username]), []byte(password)) == 1
	}
}

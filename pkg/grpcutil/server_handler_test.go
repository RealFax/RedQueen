package grpcutil_test

import (
	"context"
	"github.com/RealFax/RedQueen/pkg/grpcutil"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"testing"
)

func TestBasicAuthUnary_Success(t *testing.T) {
	authFunc := grpcutil.NewMemoryBasicAuthFunc(map[string]string{
		"username": "password",
	})

	basicAuth := grpcutil.NewBasicAuth(authFunc)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(grpcutil.MetadataAuthorization, "dXNlcm5hbWU6cGFzc3dvcmQ="))

	handler := func(ctx context.Context, req any) (any, error) {
		return "success", nil
	}

	result, err := basicAuth.Unary(ctx, nil, nil, handler)

	assert.NoError(t, err)
	assert.Equal(t, "success", result.(string))
}

func TestBasicAuthUnary_Failure(t *testing.T) {
	authFunc := grpcutil.NewMemoryBasicAuthFunc(map[string]string{
		"username": "password",
	})

	basicAuth := grpcutil.NewBasicAuth(authFunc)

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs(grpcutil.MetadataAuthorization, "invalid_token"))

	handler := func(ctx context.Context, req any) (any, error) {
		return "success", nil
	}

	result, err := basicAuth.Unary(ctx, nil, nil, handler)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestBasicAuthUnary_MissingMetadata(t *testing.T) {
	authFunc := grpcutil.NewMemoryBasicAuthFunc(map[string]string{
		"username": "password",
	})

	basicAuth := grpcutil.NewBasicAuth(authFunc)

	ctx := context.Background()

	handler := func(ctx context.Context, req any) (any, error) {
		return "success", nil
	}

	// 在 BasicAuth 中进行认证
	result, err := basicAuth.Unary(ctx, nil, nil, handler)

	// 验证结果
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

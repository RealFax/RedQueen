package client

import "sync/atomic"

var grpcPoolSize int64 = 16

func GrpcPoolSize() int64 {
	return atomic.LoadInt64(&grpcPoolSize)
}

func SetGrpcPoolSize(size int64) {
	atomic.StoreInt64(&grpcPoolSize, size)
}

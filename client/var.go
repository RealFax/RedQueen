package client

import "sync/atomic"

var maxOpenConn int64 = 16

func MaxOpenConn() int64 {
	return atomic.LoadInt64(&maxOpenConn)
}

func SetMaxOpenConn(size int64) {
	atomic.StoreInt64(&maxOpenConn, size)
}

package client

import (
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func ignoreBytes(s []byte) bool {
	return s == nil
}

func newEmptyValue[T any]() T {
	var v T
	return v
}

func newClientCall[T any](writeable bool, conn Conn, newFunc func(grpc.ClientConnInterface) T) (T, error) {
	gConn, err := func() (*grpc.ClientConn, error) {
		if writeable {
			return conn.WriteOnly()
		}
		r, err := conn.ReadOnly()
		if err != nil {
			return conn.WriteOnly()
		}
		return r, nil
	}()
	if err != nil {
		return newEmptyValue[T](), err
	}
	return newFunc(gConn), nil
}

func LockID() string {
	return uuid.New().String()
}

func Namespace(s string) *string {
	return &s
}

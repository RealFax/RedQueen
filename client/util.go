package client

import (
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

type wrapperClient[T any] struct {
	conn     *GrpcPoolConn
	instance T
}

func newClientCall[T any](writeable bool, conn Conn, newFunc func(grpc.ClientConnInterface) T) (*wrapperClient[T], error) {
	gConn, err := func() (*GrpcPoolConn, error) {
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
		return nil, err
	}
	return &wrapperClient[T]{
		conn:     gConn,
		instance: newFunc(gConn),
	}, nil
}

func LockID() string {
	return uuid.New().String()
}

func Namespace(s string) *string {
	return &s
}

func NewLeaderMonitorReceiver() *chan bool {
	c := make(chan bool, 1)
	return &c
}

package client

import (
	"encoding/hex"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"unicode/utf8"
)

type wrapperClient[T any] struct {
	conn     *grpc.ClientConn
	instance T
}

func newClientCall[T any, FC func(connInterface grpc.ClientConnInterface) T](writeable bool, conn Conn, newFunc FC) (*wrapperClient[T], error) {
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

func BString(b []byte) string {
	if utf8.Valid(b) {
		return string(b)
	}
	return hex.EncodeToString(b)
}

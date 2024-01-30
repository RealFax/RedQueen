package httputil

import (
	"github.com/RealFax/RedQueen/pkg/expr"
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"sync"
	"sync/atomic"
)

var (
	traceID atomic.Bool
	anyPool = sync.Pool{New: func() any {
		return &Caller[any]{
			release: true,
		}
	}}
)

func EnableTraceID() {
	traceID.Store(true)
}

func DisableTraceID() {
	traceID.Store(false)
}

func init() {
	// trace-id default enabled
	EnableTraceID()
}

type Ack[T any] struct {
	Status  int    `json:"status"`
	Message string `json:"msg,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type Caller[T any] struct {
	release    bool
	statusCode int
	traceID    string
	ack        Ack[T]
}

func (c *Caller[T]) reset() {
	c.statusCode = 0
	c.ack.Status = 0
	c.ack.Message = ""
	c.traceID = TraceID()
	c.ack.Data = expr.Zero[T]()
}

func (c *Caller[T]) write(w http.ResponseWriter) {
	w.WriteHeader(c.statusCode)
	w.Header().Set("Content-Type", "application/json")
	if c.traceID != "" {
		w.Header().Add("Traceparent", c.traceID)
	}

	_ = jsoniter.ConfigFastest.NewEncoder(w).Encode(c.ack)

	if c.release {
		c.reset()
		anyPool.Put(c)
	}
}

func (c *Caller[T]) GetTraceID() string {
	return c.traceID
}

func (c *Caller[T]) TraceID(i string) {
	c.traceID = i
}

func (c *Caller[T]) Message(m string) *Caller[T] {
	c.ack.Message = m
	return c
}

func (c *Caller[T]) Data(v T) *Caller[T] {
	c.ack.Data = v
	return c
}

func (c *Caller[T]) Ok(w http.ResponseWriter) { c.write(w) }

func NewAck[T any](statusCode, code int) *Caller[T] {
	return &Caller[T]{
		statusCode: statusCode,
		traceID:    TraceID(),
		ack: Ack[T]{
			Status: code,
		},
	}
}

func Any(statusCode, code int) *Caller[any] {
	r := anyPool.Get().(*Caller[any])
	r.statusCode = statusCode
	r.ack.Status = code
	return r
}

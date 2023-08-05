package red

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"sync"
	"sync/atomic"
	"time"

	orderMap "github.com/RealFax/order-map"
	"github.com/cespare/xxhash"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"

	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/internal/collapsar"
)

var (
	ErrApplyLogTimeTravelDone = errors.New("raft apply log time-travel done")
	ErrApplyLogDone           = errors.New("raft apply log done")
)

func RaftLogPayloadKey(m *serverpb.RaftLogPayload) uint64 {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer buf.Reset()

	p := make([]byte, 4)
	binary.LittleEndian.PutUint32(p, uint32(m.Command))
	buf.Write(p)

	buf.Write(m.Key)
	if m.Namespace != nil {
		buf.WriteString(*m.Namespace)
	}

	x := xxhash.New()
	io.Copy(x, buf)
	bufferPool.Put(buf)

	return x.Sum64()
}

type RaftApply interface {
	Apply(ctx *context.Context, m *serverpb.RaftLogPayload, timeout time.Duration) error
}

type ApplyFunc func(cmd []byte, timeout time.Duration) raft.ApplyFuture

type raftSingleLogApply struct {
	ApplyFunc
}

func (a *raftSingleLogApply) Apply(_ *context.Context, m *serverpb.RaftLogPayload, timeout time.Duration) error {
	b := LogPackHeader(SingleLogPack)
	cmd, err := proto.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "marshal raft log error")
	}
	b = append(b, cmd...)

	resp := a.ApplyFunc(b, timeout)
	if resp.Error() != nil {
		return resp.Error()
	}
	return ErrApplyLogDone
}

func NewRaftSingeLogApply(af ApplyFunc) RaftApply {
	return &raftSingleLogApply{af}
}

type multipleLogApplyTracker struct {
	ctx         context.Context
	cancelCause context.CancelCauseFunc
	m           *serverpb.RaftLogPayload
	// encoding message
	em []byte
}

type raftMultipleLogApply struct {
	counter                int32 // [ATOMIC]
	maxLimit               int32
	deadline, applyTimeout time.Duration
	ctx                    context.Context
	applyFunc              ApplyFunc
	onMerge                chan struct{}
	rwm                    sync.RWMutex
	filter                 orderMap.Map[uint64, *multipleLogApplyTracker]
}

func (a *raftMultipleLogApply) reset() {
	a.rwm.Lock()
	atomic.StoreInt32(&a.counter, 0)
	a.filter.Reset()
	a.rwm.Unlock()
}

func (a *raftMultipleLogApply) merge() {
	a.rwm.Lock() // stop recv apply request
	var (
		off    = 0
		size   = a.filter.Size()
		w      = collapsar.NewWriter(size)
		notify = make([]context.CancelCauseFunc, size)
	)
	a.filter.Range(func(_ uint64, value *multipleLogApplyTracker) bool {
		notify[off] = value.cancelCause
		w.Add(value.em)
		off++
		return true
	})
	a.rwm.Unlock()

	// reset state with start recv apply request
	a.reset()

	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf)
	defer buf.Reset()

	buf.Write(LogPackHeader(MultipleLogPack))
	w.Encode(buf)

	// apply log to followers
	var (
		err  = ErrApplyLogDone
		resp = a.applyFunc(buf.Bytes(), a.applyTimeout)
	)

	if resp.Error() != nil {
		err = resp.Error()
	}

	// response
	for _, causeFunc := range notify {
		causeFunc(err)
	}
}

func (a *raftMultipleLogApply) runMerge() {
	a.onMerge <- struct{}{}
}

func (a *raftMultipleLogApply) fullCounter() {
	if atomic.LoadInt32(&a.counter)+1 != atomic.LoadInt32(&a.maxLimit) {
		return
	}
	a.runMerge()
}

func (a *raftMultipleLogApply) listen() {
	ticker := time.NewTicker(a.deadline)
	defer ticker.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.runMerge()
		case <-a.onMerge:
			if atomic.LoadInt32(&a.counter) == 0 {
				continue
			}
			go a.merge()
		}
	}
}

func (a *raftMultipleLogApply) Apply(ctx *context.Context, m *serverpb.RaftLogPayload, _ time.Duration) (err error) {
	c, cancelCause := context.WithCancelCause(*ctx)
	*ctx = c
	tracker := &multipleLogApplyTracker{ctx: c, cancelCause: cancelCause, m: m}
	// preprocessing proto message
	if tracker.em, err = proto.Marshal(m); err != nil {
		return errors.Wrap(err, "marshal raft log error")
	}

	key := RaftLogPayloadKey(m)

	a.rwm.RLock()
	val, ok := a.filter.Load(key)
	if ok {
		a.filter.Delete(key)
		// time-travel close consumer context
		val.cancelCause(ErrApplyLogTimeTravelDone)
	} else {
		atomic.AddInt32(&a.counter, 1)
	}
	a.filter.Store(key, tracker)
	a.rwm.RUnlock()

	a.fullCounter()
	return nil
}

func NewRaftMultipleLogApply(ctx context.Context, maxLimit int32, deadline, applyTimeout time.Duration, af ApplyFunc) RaftApply {
	m := &raftMultipleLogApply{
		maxLimit:     maxLimit,
		deadline:     deadline,
		applyTimeout: applyTimeout,
		ctx:          ctx,
		applyFunc:    af,
		onMerge:      make(chan struct{}, 1),
		rwm:          sync.RWMutex{},
		filter:       *orderMap.New[uint64, *multipleLogApplyTracker](),
	}

	go m.listen()

	return m
}

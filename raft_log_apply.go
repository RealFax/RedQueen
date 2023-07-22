package red

import (
	"context"
	"encoding/binary"
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/internal/collapsar"
	orderMap "github.com/RealFax/order-map"
	"github.com/cespare/xxhash"
	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrApplyLogTimeTravelDone = errors.New("raft apply log time-travel done")
)

func RaftLogPayloadKey(m *serverpb.RaftLogPayload) uint64 {
	buf := bufferPool.Alloc()

	p := make([]byte, 4)
	binary.LittleEndian.PutUint32(p, uint32(m.Command))
	buf.Val().Write(p)

	buf.Val().Write(m.Key)
	if m.Namespace != nil {
		buf.Val().WriteString(*m.Namespace)
	}

	x := xxhash.New()
	io.Copy(x, buf.Val())
	buf.Free()

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
	return a.ApplyFunc(b, timeout).Error()
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
		size   = atomic.LoadInt32(&a.counter)
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

	buf := bufferPool.Alloc()
	defer buf.Free()

	buf.Val().Write(LogPackHeader(MultipleLogPack))
	w.Encode(buf.Val())

	// apply log to followers
	resp := a.applyFunc(buf.Val().Bytes(), a.applyTimeout)
	if resp == nil {
		return
	}

	for _, causeFunc := range notify {
		causeFunc(resp.Error())
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

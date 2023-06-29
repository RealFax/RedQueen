package red

import (
	"encoding/binary"
	"github.com/RealFax/RedQueen/api/serverpb"
	"github.com/RealFax/RedQueen/internal/collapsar"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"io"
)

const (
	SingleLogPack uint32 = iota
	MultipleLogPack
)

func putLogPackHeader(w io.Writer, typ uint32) {
	p := make([]byte, 4)
	binary.LittleEndian.PutUint32(p, typ)
	w.Write(p)
}

func getLogPackHeader(r io.Reader) uint32 {
	p := make([]byte, 4)
	r.Read(p)
	return binary.LittleEndian.Uint32(p)
}

func PackSingleLog(w io.Writer, m *serverpb.RaftLogPayload) error {
	putLogPackHeader(w, SingleLogPack)
	cmd, err := proto.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "marshal raft log error")
	}
	w.Write(cmd)
	return nil
}

func UnpackSingleLog(r io.Reader) (*serverpb.RaftLogPayload, error) {
	if getLogPackHeader(r) != SingleLogPack {
		return nil, errors.New("invalid single log header")
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	m := &serverpb.RaftLogPayload{}
	if err = proto.Unmarshal(b, m); err != nil {
		return nil, errors.Wrap(err, "unmarshal raft log error")
	}
	return m, nil
}

func UnpackMultipleLog(r io.Reader) ([]*serverpb.RaftLogPayload, error) {
	if getLogPackHeader(r) != MultipleLogPack {
		return nil, errors.New("invalid multiple log header")
	}
	cr, err := collapsar.NewReader(r)
	if err != nil {
		return nil, err
	}
	var (
		logs []*serverpb.RaftLogPayload
		m    *serverpb.RaftLogPayload
	)
	for {
		b, rErr := cr.Next()
		if rErr != nil {
			if rErr == io.EOF {
				return logs, nil
			}
			return nil, err
		}
		if err = proto.Unmarshal(b, m); err != nil {
			return nil, errors.Wrap(err, "unmarshal raft log error")
		}
		logs = append(logs, m)
	}
}

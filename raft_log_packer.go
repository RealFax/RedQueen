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

func LogPackHeader(typ uint32) []byte {
	p := make([]byte, 4)
	binary.LittleEndian.PutUint32(p, typ)
	return p
}

func GetLogPackHeader(r io.Reader) uint32 {
	p := make([]byte, 4)
	r.Read(p)
	return binary.LittleEndian.Uint32(p)
}

func unpackSingleLog(r io.Reader) (*serverpb.RaftLogPayload, error) {
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

func unpackMultipleLog(r io.Reader) ([]*serverpb.RaftLogPayload, error) {
	cr, err := collapsar.NewReader(r)
	if err != nil {
		return nil, err
	}
	var (
		logs []*serverpb.RaftLogPayload
		m    serverpb.RaftLogPayload
	)
	for {
		b, rErr := cr.Next()
		if rErr != nil {
			if rErr == io.EOF {
				return logs, nil
			}
			return nil, err
		}
		if err = proto.Unmarshal(b, &m); err != nil {
			return nil, errors.Wrap(err, "unmarshal raft log error")
		}
		logs = append(logs, &m)
	}
}

func UnpackLog(r io.Reader) ([]*serverpb.RaftLogPayload, error) {
	switch GetLogPackHeader(r) {
	case SingleLogPack:
		m, err := unpackSingleLog(r)
		if err != nil {
			return nil, err
		}
		return []*serverpb.RaftLogPayload{m}, nil
	case MultipleLogPack:
		logs, err := unpackMultipleLog(r)
		if err != nil {
			return nil, err
		}
		return logs, nil
	default:
		return nil, errors.New("auto unmarshal raft log error, no method")
	}
}

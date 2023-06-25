package red

import (
	"encoding/binary"
	"github.com/RealFax/RedQueen/api/serverpb"
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

func NewSingleLogPack(w io.Writer, m *serverpb.RaftLogPayload) error {
	putLogPackHeader(w, SingleLogPack)
	cmd, err := proto.Marshal(m)
	if err != nil {
		return errors.Wrap(err, "marshal raft log error")
	}
	w.Write(cmd)
	return nil
}

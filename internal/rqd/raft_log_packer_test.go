package rqd_test

import (
	"bytes"
	red "github.com/RealFax/RedQueen/internal/rqd"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLogPackHeader(t *testing.T) {
	header := red.LogPackHeader(red.SingleLogPack)
	assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, header)

	header = red.LogPackHeader(red.MultipleLogPack)
	assert.Equal(t, []byte{0x01, 0x00, 0x00, 0x00}, header)
}

func TestGetLogPackHeader(t *testing.T) {
	buf := bytes.NewReader([]byte{0x00, 0x00, 0x00, 0x00})
	typ := red.GetLogPackHeader(buf)
	assert.Equal(t, uint32(0x00), typ)

	buf = bytes.NewReader([]byte{0x01, 0x00, 0x00, 0x00})
	typ = red.GetLogPackHeader(buf)
	assert.Equal(t, uint32(0x01), typ)
}

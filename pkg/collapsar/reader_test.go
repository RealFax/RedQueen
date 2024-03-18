package collapsar_test

import (
	"bytes"
	"github.com/RealFax/RedQueen/pkg/collapsar"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

var (
	matrix = [][]byte{
		{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9},
	}
)

func TestReader_Next(t *testing.T) {
	w := collapsar.NewWriter(10)
	for _, v := range matrix {
		assert.NoError(t, w.Add(v))
	}

	buf := &bytes.Buffer{}
	assert.NoError(t, w.Encode(buf))

	r, err := collapsar.NewReader(buf)
	assert.NoError(t, err)

	var (
		cur int
		p   []byte
	)
	for {
		if p, err = r.Next(); err != nil {
			if err == io.EOF {
				return
			}
			assert.NoError(t, err)
		}
		assert.Equal(t, matrix[cur], p)
		cur++
	}

}

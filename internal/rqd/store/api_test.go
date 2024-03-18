package store_test

import (
	"github.com/RealFax/RedQueen/internal/rqd/store"
	"github.com/RealFax/RedQueen/pkg/expr"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWatchValue_Deleted(t *testing.T) {
	var (
		value1 = store.WatchValue{}
		value2 = store.WatchValue{Value: expr.Pointer([]byte("Test"))}
	)

	assert.True(t, value1.Deleted())
	assert.False(t, value2.Deleted())
}

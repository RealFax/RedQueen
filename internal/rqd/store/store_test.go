package store_test

import (
	"github.com/RealFax/RedQueen/internal/rqd/store"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnwrapGet(t *testing.T) {
	value1 := &store.Value{}
	value2 := &store.Value{Data: []byte("Test")}

	_value1, err := store.UnwrapGet(value1, nil)
	assert.NoError(t, err)
	assert.Nil(t, _value1)

	_value2, err := store.UnwrapGet(value2, nil)
	assert.NoError(t, err)
	assert.NotNil(t, _value2)

	_value3, err := store.UnwrapGet(value1, errors.New(""))
	assert.Error(t, err)
	assert.Nil(t, _value3)

	_value4, err := store.UnwrapGet(value2, errors.New(""))
	assert.Error(t, err)
	assert.NotNil(t, _value4)
}

package utils_test

import (
	"github.com/RealFax/RedQueen/utils"
	"testing"
)

func TestString2Bytes(t *testing.T) {
	t.Log(utils.String2Bytes("Hello, world"))
}

func TestBytes2String(t *testing.T) {
	t.Log(utils.Bytes2String([]byte("Hello, world")))
}

package hack_test

import (
	"testing"

	"github.com/RealFax/RedQueen/common/hack"
)

func TestString2Bytes(t *testing.T) {
	t.Log(hack.String2Bytes("Hello, world"))
}

func TestBytes2String(t *testing.T) {
	t.Log(hack.Bytes2String([]byte("Hello, world")))
}

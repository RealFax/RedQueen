package memory

import (
	"crypto/rand"
	"encoding/binary"
)

var (
	seed uint64
)

func init() {
	// init sum seed
	seedBytes := make([]byte, 8)
	rand.Read(seedBytes)
	seed = binary.LittleEndian.Uint64(seedBytes)
}

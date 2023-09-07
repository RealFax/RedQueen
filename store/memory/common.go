package memory

import "github.com/OneOfOne/xxhash"

func KeySum(key []byte) uint64 {
	return xxhash.Checksum64S(key, seed)
}

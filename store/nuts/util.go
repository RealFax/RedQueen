package nuts

// WatchKey
// from: https://en.wikipedia.org/wiki/Jenkins_hash_function (Jenkins' One-At-A-Time hashing)
func WatchKey(key []byte) (hash uint64) {
	for i := 0; i < len(key); i++ {
		hash += (uint64)(key[i])
		hash += hash << 10
		hash ^= hash >> 6
	}
	hash += hash << 3
	hash ^= hash >> 1
	hash += hash << 15
	return
}

package nuts

import (
	"archive/tar"
	"compress/gzip"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nutsdb/nutsdb"
)

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

func PrefixKey(prefix []byte) string {
	return base64.StdEncoding.EncodeToString(prefix)
}

func BackupReader(dst string, src io.Reader) error {
	{
		ugz, err := gzip.NewReader(src)
		if err != nil {
			return err
		}
		src = ugz
	}
	reader := tar.NewReader(src)
	rootDir := ""
	for {
		header, err := reader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// check if the path is vulnerable
		if strings.Contains(header.FileInfo().Name(), "..") {
			continue
		}

		path := filepath.Join(dst, header.Name)

		// handle dir
		if header.FileInfo().IsDir() {
			if rootDir == "" {
				rootDir = header.FileInfo().Name()
				if err = os.MkdirAll(dst, header.FileInfo().Mode()); err != nil {
					return err
				}
				continue
			}
			if err = os.MkdirAll(path, header.FileInfo().Mode()); err != nil {
				return err
			}
			continue
		}

		path = filepath.Clean(strings.Replace(path, rootDir, "", 1))
		// handle file
		f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, header.FileInfo().Mode())
		if err != nil {
			return err
		}
		if _, err = io.Copy(f, reader); err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()
	}
	return nil
}

func ReadTTL(md *nutsdb.MetaData) uint32 {
	return uint32((md.Timestamp / 1000) + uint64(md.TTL) - uint64(time.Now().Unix()))
}

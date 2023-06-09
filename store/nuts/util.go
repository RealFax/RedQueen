package nuts

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"strings"
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
			f.Close()
			return err
		}
		f.Close()
	}
	return nil
}

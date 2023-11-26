package nuts

import (
	"archive/tar"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
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

func BackupWriter(src string, dst io.Writer) error {
	gzipWriter := gzip.NewWriter(dst)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Update the header name to use relative paths
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If the file is a regular file, write its contents to the tarball
		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(tarWriter, file)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func BackupReader(dst string, src io.Reader) error {
	gzipReader, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return err
		}

		targetPath := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported tar entry: %s", header.Name)
		}
	}

	return nil
}

func ReadTTL(md *nutsdb.MetaData) uint32 {
	return uint32((md.Timestamp / 1000) + uint64(md.TTL) - uint64(time.Now().Unix()))
}

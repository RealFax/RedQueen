package fs

import (
	"io"
	"os"
	"strings"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

func MustOpenWithFlag(path string, flag int) (*os.File, error) {
	dstDir := path[0 : strings.LastIndex(path, "/")+1]
	if !IsExist(dstDir) {
		var err error
		if err = os.MkdirAll(dstDir, os.ModePerm); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(path, flag, 0644)
}

// MustOpen if the file does not exist, create a file and open it in overwrite mode
func MustOpen(path string) (*os.File, error) {
	return MustOpenWithFlag(path, os.O_WRONLY|os.O_TRUNC|os.O_CREATE)
}

func MustOpenWithAppend(path string) (*os.File, error) {
	return MustOpenWithFlag(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE)
}

func GetFilePerm(path string) (string, error) {
	f, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	return f.Mode().Perm().String(), nil
}

// WriteAsBytes will overwrite the target file
func WriteAsBytes(path string, b []byte) error {
	f, err := MustOpen(path)
	if err != nil {
		return err
	}
	defer f.Close()

	n, _ := f.Seek(io.SeekStart, io.SeekEnd)
	_, err = f.WriteAt(b, n)
	return err
}

func ReadAsBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

package fs

import (
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

package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
)

func CopyDir(scrDir, dest string) error {
	entries, err := os.ReadDir(scrDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		var (
			fileInfo os.FileInfo

			srcPath = filepath.Join(scrDir, entry.Name())
			dstPath = filepath.Join(dest, entry.Name())
		)

		if fileInfo, err = os.Stat(srcPath); err != nil {
			return err
		}

		stat, ok := fileInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return fmt.Errorf("failed to get raw syscall.Stat_t data for '%s'", srcPath)
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err = CreateIfNotExists(dstPath, 0755); err != nil {
				return err
			}
			if err = CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err = CopySymLink(srcPath, dstPath); err != nil {
				return err
			}
		default:
			if err = Copy(srcPath, dstPath); err != nil {
				return err
			}
		}

		if err = os.Lchown(dstPath, int(stat.Uid), int(stat.Gid)); err != nil {
			return err
		}

		if fileInfo, err = entry.Info(); err != nil {
			return err
		}

		isSymlink := fileInfo.Mode()&os.ModeSymlink != 0
		if !isSymlink {
			if err = os.Chmod(dstPath, fileInfo.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

func Copy(srcFile, dstFile string) error {
	out, err := MustOpen(dstFile)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}

	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func CreateIfNotExists(dir string, perm os.FileMode) error {
	if IsExist(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

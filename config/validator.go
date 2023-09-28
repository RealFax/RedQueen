package config

import (
	"os"

	"github.com/pkg/errors"
)

type Validator interface {
	Valid() error
}

type EnumStoreBackend string

const (
	StoreBackendNuts EnumStoreBackend = "nuts"
)

func (b EnumStoreBackend) Valid() error {
	switch b {
	case StoreBackendNuts:
		return nil
	default:
		return errors.New("unknown store backend type")
	}
}

type EnumNutsRWMode string

const (
	NutsRWModeFileIO EnumNutsRWMode = "fileio"
	NutsRWModeMMap   EnumNutsRWMode = "mmap"
)

func (m EnumNutsRWMode) Valid() error {
	switch m {
	case NutsRWModeFileIO, NutsRWModeMMap:
		return nil
	default:
		return errors.New("unknown nuts rw mode")
	}
}

type EnumLogLogger string

const (
	LogLoggerZap      EnumLogLogger = "zap"
	LogLoggerInternal EnumLogLogger = "internal"
)

func (l EnumLogLogger) Valid() error {
	switch l {
	case LogLoggerZap, LogLoggerInternal:
		return nil
	default:
		return errors.New("unknown log logger type")
	}
}

type FilePath string

func (p FilePath) Valid() error {
	_, err := os.Stat(string(p))
	if err == nil || os.IsExist(err) {
		return nil
	}
	return errors.New("file/dir not found")
}

type stringValidator interface {
	EnumStoreBackend | EnumNutsRWMode | EnumLogLogger | FilePath
}

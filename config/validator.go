package config

import (
	"github.com/pkg/errors"
	"os"
)

type Validator interface {
	Valid() error
}

type EnumClusterState string

const (
	ClusterStateNew      EnumClusterState = "new"
	ClusterStateExisting EnumClusterState = "existing"
)

func (s EnumClusterState) Valid() error {
	switch s {
	case ClusterStateNew, ClusterStateExisting:
		return nil
	default:
		return errors.New("unknown cluster state type")
	}
}

type EnumLogLogger string

const (
	LogLoggerZap      EnumLogLogger = "zap"
	LogLoggerCapnslog EnumLogLogger = "capnslog"
)

func (l EnumLogLogger) Valid() error {
	switch l {
	case LogLoggerZap, LogLoggerCapnslog:
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

func Validate() {

}

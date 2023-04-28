package locker

import (
	"github.com/pkg/errors"
	"time"
)

const (
	Namespace string = "LockerMutex"
	Deadline         = time.Minute * 10
)

var (
	ErrStatusBusy = errors.New("status busy")
)

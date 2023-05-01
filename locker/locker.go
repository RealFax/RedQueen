package locker

import (
	"github.com/pkg/errors"
	"time"
)

const (
	Namespace string = "Locker"
	Deadline         = time.Second * 30
)

var (
	ErrStatusBusy = errors.New("status busy")
)

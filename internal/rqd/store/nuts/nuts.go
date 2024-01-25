package nuts

import (
	"github.com/RealFax/RedQueen/internal/rqd/store"
	"sync"
	"sync/atomic"

	"github.com/nutsdb/nutsdb"
	"github.com/pkg/errors"
)

// db state

const (
	StateOk uint32 = iota
	StateBreak
)

var (
	KeyInitBucket = []byte("_init_key")
	ErrStateBreak = errors.New("state break")
)

// RWMode represents the read and write mode.
type RWMode = nutsdb.RWMode

const (
	// FileIO represents the read and write mode using standard I/O.
	FileIO = nutsdb.FileIO

	// MMap represents the read and write mode using mmap.
	MMap = nutsdb.MMap
)

var strictMode atomic.Bool

func EnableStrictMode()  { strictMode.Store(true) }
func DisableStrictMode() { strictMode.Store(false) }

func init() {
	EnableStrictMode()
}

type Config struct {
	NodeNum int64
	Sync    bool
	DataDir string
	RWMode  RWMode
}

type DB struct {
	state *uint32 // atomic

	db      *atomic.Pointer[nutsdb.DB]
	options []nutsdb.Option

	// root watcher
	watcher *Watcher

	// watcherChild for the current namespace
	watcherChild *WatcherChild

	mu        sync.RWMutex
	namespace string
	dataDir   string
}

func New(cfg Config) (store.Store, error) {
	opts := []nutsdb.Option{
		nutsdb.WithDir(cfg.DataDir),
		nutsdb.WithSyncEnable(cfg.Sync),
		nutsdb.WithNodeNum(cfg.NodeNum),
		nutsdb.WithRWMode(cfg.RWMode),
		// nutsdb.WithSegmentSize(128 * nutsdb.MB),
	}
	db, err := nutsdb.Open(nutsdb.DefaultOptions, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "can't create nuts store api")
	}

	rootWatcher := &Watcher{}

	dbPtr := atomic.Pointer[nutsdb.DB]{}
	dbPtr.Store(db)

	return &DB{
		state:        new(uint32),
		db:           &dbPtr,
		options:      opts,
		watcher:      rootWatcher,
		watcherChild: rootWatcher.UseTarget(store.DefaultNamespace),
		namespace:    store.DefaultNamespace,
		dataDir:      cfg.DataDir,
	}, nil
}

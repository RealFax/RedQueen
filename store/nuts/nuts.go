package nuts

import (
	"github.com/RealFax/RedQueen/store"
	"github.com/nutsdb/nutsdb"
	"github.com/pkg/errors"
	"sync"
	"sync/atomic"
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
}

type storeAPI struct {
	db *nutsdb.DB

	// root watcher
	watcher *Watcher

	// watcherChild for the current namespace
	watcherChild *WatcherChild

	mu        sync.Mutex
	namespace string
}

func New(cfg Config) (store.Store, error) {
	db, err := nutsdb.Open(
		nutsdb.DefaultOptions,
		nutsdb.WithDir(cfg.DataDir),
		nutsdb.WithSyncEnable(cfg.Sync),
		nutsdb.WithNodeNum(cfg.NodeNum),
	)
	if err != nil {
		return nil, errors.Wrap(err, "can't create nuts store api")
	}

	rootWatcher := &Watcher{}

	return &storeAPI{
		db:           db,
		watcher:      rootWatcher,
		watcherChild: rootWatcher.Namespace(store.DefaultNamespace),
		namespace:    store.DefaultNamespace,
	}, nil
}

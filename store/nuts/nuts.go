package nuts

import (
	"github.com/RealFax/RedQueen/store"
	"github.com/nutsdb/nutsdb"
	"github.com/pkg/errors"
	"sync"
)

const DefaultNamespace = "RED_QUEEN"

type Config struct {
	NodeNum int64
	Sync    bool
	DataDir string
}

type storeAPI struct {
	db        *nutsdb.DB
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
		return nil, errors.Wrap(err, "can't create store session")
	}
	return &storeAPI{
		db:        db,
		namespace: DefaultNamespace,
	}, nil
}

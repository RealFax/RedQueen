package RedQueen

import (
	"errors"
	"github.com/RealFax/RedQueen/config"
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/store/nuts"
)

func newNutsStore(cfg config.Store) (store.Store, error) {
	if cfg.Nuts.StrictMode {
		nuts.EnableStrictMode()
	} else {
		nuts.DisableStrictMode()
	}

	return nuts.New(nuts.Config{
		NodeNum: cfg.Nuts.NodeNum,
		Sync:    cfg.Nuts.Sync,
		DataDir: cfg.Nuts.DataDir,
	})
}

func newStoreBackend(cfg config.Store) (store.Store, error) {
	handle, ok := map[config.EnumStoreBackend]func(config.Store) (store.Store, error){
		config.StoreBackendNuts: newNutsStore,
	}[cfg.Backend]
	if !ok {
		return nil, errors.New("unsupported store backend")
	}
	return handle(cfg)
}

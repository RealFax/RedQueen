package rqd

import (
	"errors"
	config2 "github.com/RealFax/RedQueen/internal/rqd/config"
	"github.com/RealFax/RedQueen/internal/rqd/store"
	"github.com/RealFax/RedQueen/internal/rqd/store/nuts"
	"path/filepath"
)

func newNutsStore(cfg config2.Store, dir string) (store.Store, error) {
	if cfg.Nuts.StrictMode {
		nuts.EnableStrictMode()
	} else {
		nuts.DisableStrictMode()
	}

	return nuts.New(nuts.Config{
		NodeNum: cfg.Nuts.NodeNum,
		Sync:    cfg.Nuts.Sync,
		DataDir: filepath.Join(dir, StoreSuffix),
		RWMode: func() nuts.RWMode {
			switch cfg.Nuts.RWMode {
			case config2.NutsRWModeFileIO:
				return nuts.FileIO
			case config2.NutsRWModeMMap:
				return nuts.MMap
			default:
				return nuts.FileIO
			}
		}(),
	})
}

func newStoreBackend(cfg config2.Store, dir string) (store.Store, error) {
	handle, ok := map[config2.EnumStoreBackend]func(config2.Store, string) (store.Store, error){
		config2.StoreBackendNuts: newNutsStore,
	}[cfg.Backend]
	if !ok {
		return nil, errors.New("unsupported actions backend")
	}
	return handle(cfg, dir)
}

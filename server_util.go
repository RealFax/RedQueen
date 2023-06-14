package red

import (
	"errors"
	"github.com/RealFax/RedQueen/config"
	"github.com/RealFax/RedQueen/store"
	"github.com/RealFax/RedQueen/store/nuts"
	"path/filepath"
)

func newNutsStore(cfg config.Store, dir string) (store.Store, error) {
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
			case config.NutsRWModeFileIO:
				return nuts.FileIO
			case config.NutsRWModeMMap:
				return nuts.MMap
			default:
				return nuts.FileIO
			}
		}(),
	})
}

func newStoreBackend(cfg config.Store, dir string) (store.Store, error) {
	handle, ok := map[config.EnumStoreBackend]func(config.Store, string) (store.Store, error){
		config.StoreBackendNuts: newNutsStore,
	}[cfg.Backend]
	if !ok {
		return nil, errors.New("unsupported store backend")
	}
	return handle(cfg, dir)
}

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
		DataDir: filepath.Join(dir, "fsm"),
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

//func loadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
//	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
//	if err != nil {
//		return nil, err
//	}
//	return &tls.Config{
//		Certificates:     []tls.Certificate{cert},
//		MinVersion:       tls.VersionTLS12,
//		CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
//		CipherSuites: []uint16{
//			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
//			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
//			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
//			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
//		},
//	}, nil
//}

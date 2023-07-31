package config

const DefaultConfigPath string = "./config.toml"

// -- node default value

const (
	DefaultNodeDataDir          string = "./data"
	DefaultNodeListenPeerAddr   string = "127.0.0.1:5290"
	DefaultNodeListenClientAddr string = "127.0.0.1:5230"
	DefaultNodeMaxSnapshots     uint32 = 5
	DefaultNodeRequestsMerged   bool   = false
)

// -- store default value

const (
	DefaultStoreBackend = string(StoreBackendNuts)

	DefaultStoreNutsNodeNum    int64  = 1
	DefaultStoreNutsSync       bool   = false
	DefaultStoreNutsStrictMode bool   = false
	DefaultStoreNutsRWMode     string = "fileio"
)

// -- cluster default value

const (
	DefaultClusterState string = string(ClusterStateNew)
)

// -- log default value

const (
	DefaultLogLogger string = string(LogLoggerZap)
)

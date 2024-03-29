package config

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/RealFax/RedQueen/pkg/fs"
	"github.com/pkg/errors"
	"os"
	"path"
)

type ServerEnv interface {
	FirstRun() bool
	ConfigFile() string
}

type env struct {
	firstRun     bool
	initLockFile string

	configFile string
}

func (r *env) FirstRun() bool {
	return r.firstRun
}

func (r *env) ConfigFile() string {
	return r.configFile
}

type NodeTLS struct {
	Auto     bool   `toml:"auto"`
	CertFile string `toml:"cert-file"`
	KeyFile  string `toml:"key-file"`
}

type Node struct {
	ID               string  `toml:"id"`
	DataDir          string  `toml:"data-dir"`
	ListenPeerAddr   string  `toml:"listen-peer-addr"`
	ListenClientAddr string  `toml:"listen-client-addr"`
	ListenHttpAddr   string  `toml:"listen-http-addr"`
	MaxSnapshots     uint32  `toml:"max-snapshots"`
	RequestsMerged   bool    `toml:"requests-merged"`
	TLS              NodeTLS `toml:"tls"`
}

type StoreNuts struct {
	NodeNum    int64          `toml:"node-num"`
	Sync       bool           `toml:"sync"`
	StrictMode bool           `toml:"strict-mode"`
	RWMode     EnumNutsRWMode `toml:"rw-mode"`
}

type Store struct {
	Backend EnumStoreBackend `toml:"backend"`
	Nuts    StoreNuts        `toml:"nuts"`
}

type ClusterBootstrap struct {
	Name     string `toml:"name"`
	PeerAddr string `toml:"peer-addr"`
}

type Cluster struct {
	Bootstrap []ClusterBootstrap `toml:"bootstrap"`
}

type Misc struct {
	PPROF bool `toml:"pprof"`
}

type BasicAuth map[string]string

type Config struct {
	*env
	Node      `toml:"node"`
	Store     `toml:"store"`
	Cluster   `toml:"cluster"`
	Misc      `toml:"misc"`
	BasicAuth `toml:"basic-auth"`
}

func (c *Config) setupEnv() {
	c.env.initLockFile = path.Join(c.Node.DataDir, ".init.lock")

	// write init lock
	if !fs.IsExist(c.env.initLockFile) {
		c.env.firstRun = true

		f, err := fs.MustOpen(c.env.initLockFile)
		if err != nil {
			panic("Setup config error:" + err.Error())
		}
		defer f.Close()
		_, _ = f.WriteString("LOCK")
	}
}

func (c *Config) Env() ServerEnv {
	return c.env
}

func newConfigEntity() *Config {
	return &Config{
		env: &env{},
	}
}

func bindServerFromArgs(cfg *Config, args ...string) error {
	if len(args) < 1 {
		return errors.New("invalid program args")
	}

	f := flag.NewFlagSet("server", flag.ExitOnError)

	f.Usage = func() {
		fmt.Fprint(f.Output(), serverUsage)
		f.PrintDefaults()
	}

	f.StringVar(&cfg.env.configFile, "config-file", "", "config file path")

	// main config::node
	f.StringVar(&cfg.Node.ID, "node-id", "", "unique node id")
	f.StringVar(&cfg.Node.DataDir, "data-dir", DefaultNodeDataDir, "path to the data dir")
	f.StringVar(&cfg.Node.ListenPeerAddr, "listen-peer-addr", DefaultNodeListenPeerAddr, "address to raft listen")
	f.StringVar(&cfg.Node.ListenClientAddr, "listen-client-addr", DefaultNodeListenClientAddr, "address to grpc listen")
	f.StringVar(&cfg.Node.ListenHttpAddr, "listen-http-addr", "", "address to http listen, if it is empty, it means that http server is not used.")
	f.Var(newUInt32Value(DefaultNodeMaxSnapshots, &cfg.Node.MaxSnapshots), "max-snapshots", "max number to snapshots(raft)")
	f.BoolVar(&cfg.Node.RequestsMerged, "requests-merged", DefaultNodeRequestsMerged, "enable raft apply log requests merged")

	// main config::node::tls
	f.BoolVar(&cfg.Node.TLS.Auto, "auto-tls", false, "auto generator tls")
	f.StringVar(&cfg.Node.TLS.CertFile, "tls-cert-file", "", "tls certificate file")
	f.StringVar(&cfg.Node.TLS.KeyFile, "tls-key-file", "", "tls key file")

	// main config::store
	f.Var(newValidatorStringValue[EnumStoreBackend](DefaultStoreBackend, &cfg.Store.Backend), "store-backend", "")

	// main config::store::nuts
	f.Int64Var(&cfg.Store.Nuts.NodeNum, "nuts-node-num", DefaultStoreNutsNodeNum, "node-id in the system")
	f.BoolVar(&cfg.Store.Nuts.Sync, "nuts-sync", DefaultStoreNutsSync, "enable sync write")
	f.BoolVar(&cfg.Store.Nuts.StrictMode, "nuts-strict-mode", DefaultStoreNutsStrictMode, "enable strict mode")
	f.Var(newValidatorStringValue[EnumNutsRWMode](DefaultStoreNutsRWMode, &cfg.Store.Nuts.RWMode), "nuts-rw-mode", "select read & write mode, options: fileio, mmap")

	// main config::cluster::bootstrap(s)
	// in cli: node-1@peer_addr,node-2@peer_addr
	f.Var(newClusterBootstrapsValue("", &cfg.Cluster.Bootstrap), "cluster-bootstrap", "bootstrap at cluster startup, e.g. : node-1@peer_addr,node-2@peer_addr")

	// main config::misc
	f.BoolVar(&cfg.Misc.PPROF, "d-pprof", false, "")

	// main config::basic-auth
	f.Var(newStringMap("", (*map[string]string)(&cfg.BasicAuth)), "basic-auth", "grpc, http api endpoint basic auth map, e.g. : root:toor,admin:123456")

	return f.Parse(args)
}

func bindServerFromEnv(cfg *Config) {
	EnvStringVar(&cfg.env.configFile, "RQ_CONFIG_FILE", "")

	// main config::node
	EnvStringVar(&cfg.Node.ID, "RQ_NODE_ID", "")
	EnvStringVar(&cfg.Node.DataDir, "RQ_DATA_DIR", DefaultNodeDataDir)
	EnvStringVar(&cfg.Node.ListenPeerAddr, "RQ_LISTEN_PEER_ADDR", DefaultNodeListenPeerAddr)
	EnvStringVar(&cfg.Node.ListenClientAddr, "RQ_LISTEN_CLIENT_ADDR", DefaultNodeListenClientAddr)
	EnvStringVar(&cfg.Node.ListenHttpAddr, "RQ_LISTEN_HTTP_ADDR", "")
	BindEnvVar(newUInt32Value(DefaultNodeMaxSnapshots, &cfg.Node.MaxSnapshots), "RQ_MAX_SNAPSHOTS")
	EnvBoolVar(&cfg.Node.RequestsMerged, "RQ_REQUESTS_MERGED", DefaultNodeRequestsMerged)

	// main config::node::tls
	EnvBoolVar(&cfg.Node.TLS.Auto, "RQ_AUTO_TLS", false)
	EnvStringVar(&cfg.Node.TLS.CertFile, "RQ_TLS_CERT_FILE", "")
	EnvStringVar(&cfg.Node.TLS.KeyFile, "RQ_TLS_KEY_FILE", "")

	// main config::store
	BindEnvVar(newValidatorStringValue[EnumStoreBackend](DefaultStoreBackend, &cfg.Store.Backend), "RQ_STORE_BACKEND")

	// main config::store::nuts
	EnvInt64Var(&cfg.Store.Nuts.NodeNum, "RQ_NUTS_NODE_NUM", DefaultStoreNutsNodeNum)
	EnvBoolVar(&cfg.Store.Nuts.Sync, "RQ_NUTS_SYNC", DefaultStoreNutsSync)
	EnvBoolVar(&cfg.Store.Nuts.StrictMode, "RQ_NUTS_STRICT_MODE", DefaultStoreNutsStrictMode)
	BindEnvVar(newValidatorStringValue[EnumNutsRWMode](DefaultStoreNutsRWMode, &cfg.Store.Nuts.RWMode), "RQ_NUTS_RW_MODE")

	// main config::cluster::bootstrap(s)
	BindEnvVar(newClusterBootstrapsValue("", &cfg.Cluster.Bootstrap), "RQ_CLUSTER_BOOTSTRAP")

	// main config::misc
	EnvBoolVar(&cfg.Misc.PPROF, "RQ_DEBUG_PPROF", false)

	// main config::basic-auth
	BindEnvVar(newStringMap("", (*map[string]string)(&cfg.BasicAuth)), "RQ_BASIC_AUTH")
}

func bindFromConfigFile(cfg *Config, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = toml.NewDecoder(f).Decode(cfg)
	return err
}

func ReadFromArgs(args ...string) (*Config, error) {
	cfg := newConfigEntity()

	args = args[1:]
	if len(args) == 0 {
		return nil, errors.New("no subcommand provided")
	}

	switch args[0] {
	case "server":
		if err := bindServerFromArgs(cfg, args[1:]...); err != nil {
			return nil, err
		}
	default:
		fmt.Fprint(os.Stderr, usage)
		return nil, errors.New("unknown subcommand")
	}

	// override config from args
	if cfg.Env().ConfigFile() != "" {
		if err := bindFromConfigFile(cfg, cfg.env.ConfigFile()); err != nil {
			return nil, err
		}
		return cfg, nil
	}

	return cfg, nil
}

func ReadFromPath(path string) (*Config, error) {
	cfg := newConfigEntity()
	if err := bindFromConfigFile(cfg, path); err != nil {
		return nil, err
	}
	return cfg, nil
}

func ReadFromEnv() *Config {
	cfg := newConfigEntity()
	bindServerFromEnv(cfg)
	return cfg
}

func New(args ...string) (cfg *Config, err error) {
	defer func() {
		if err == nil {
			cfg.setupEnv()
		}
	}()

	configPath := DefaultConfigPath

	cfg = ReadFromEnv()
	if cfg.Node.ID != "" {
		if cfg.env.configFile != "" {
			configPath = cfg.env.configFile
			goto HandleConfigFile
		}
		return
	}

	if cfg, err = ReadFromArgs(args...); err != nil {
		return
	}
	if cfg.Node.ID != "" {
		if cfg.env.configFile != "" {
			configPath = cfg.env.configFile
			goto HandleConfigFile
		}
		return
	}

HandleConfigFile:
	cfg, err = ReadFromPath(configPath)
	return
}

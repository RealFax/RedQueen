package config

import (
	"flag"

	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"os"
)

type ServerEnv interface {
	FirstRun() bool
	ConfigFile() string
}

type env struct {
	firstRun   bool
	configFile string
}

func (r env) FirstRun() bool {
	return r.firstRun
}

func (r env) ConfigFile() string {
	return r.configFile
}

type Node struct {
	ID               string `toml:"id"`
	DataDir          string `toml:"data-dir"`
	ListenPeerAddr   string `toml:"listen-peer-addr"`
	ListenClientAddr string `toml:"listen-client-addr"`
	MaxSnapshots     uint32 `toml:"max-snapshots"`
}

type StoreNuts struct {
	NodeNum    int64 `toml:"node-num"`
	Sync       bool  `toml:"sync"`
	StrictMode bool  `toml:"strict-mode"`
}

type Store struct {
	Backend EnumStoreBackend
	Nuts    StoreNuts
}

type ClusterBootstrap struct {
	Name     string `toml:"name"`
	PeerAddr string `toml:"peer-addr"`
}

type Cluster struct {
	// State enum: new, existing
	State     EnumClusterState   `toml:"state"`
	Token     string             `toml:"token"`
	Bootstrap []ClusterBootstrap `toml:"bootstrap"`
}

type Security struct {
	EnableTLS     bool `toml:"enable-tls"`
	EnablePeerTLS bool `toml:"enable-peer-tls"`
	AutoTLS       bool `toml:"auto-tls"`
	PeerAutoTLS   bool `toml:"peer-auto-tls"`

	TLSCert     string `toml:"tls-cert"`
	TLSKey      string `toml:"tls-key"`
	PeerTLSCert string `toml:"peer-tls-cert"`
	PeerTLSKey  string `toml:"peer-tls-key"`
}

type Log struct {
	Debug bool `toml:"debug"`
	// Logger enum: zap, internal
	Logger    EnumLogLogger `toml:"logger"`
	OutputDir string        `toml:"output-dir"`
}

type Misc struct {
	PPROF bool `toml:"pprof"`
}

type Auth struct {
	Token string `toml:"token"`
}

type Config struct {
	env
	Node     `toml:"node"`
	Store    `toml:"store"`
	Cluster  `toml:"cluster"`
	Security `toml:"security"`
	Log      `toml:"log"`
	Misc     `toml:"misc"`
	Auth     `toml:"auth"`
}

func (c *Config) setupEnv() {
	c.env.firstRun = c.Cluster.State == ClusterStateNew
}

func (c *Config) Env() ServerEnv {
	return c.env
}

func newConfigEntity() *Config {
	return new(Config)
}

func bindServerFromArgs(cfg *Config, args ...string) error {
	if len(args) < 1 {
		return errors.New("invalid program args")
	}

	fs := flag.NewFlagSet("server", flag.ExitOnError)

	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage of RedQueen:")
		fs.PrintDefaults()
	}

	fs.StringVar(&cfg.env.configFile, "config-file", "", "config file path")

	// main config::node
	fs.StringVar(&cfg.Node.ID, "node-id", "", "unique node id")
	fs.StringVar(&cfg.Node.DataDir, "data-dir", DefaultNodeDataDir, "path to the data dir")
	fs.StringVar(&cfg.Node.ListenPeerAddr, "listen-peer-addr", DefaultNodeListenPeerAddr, "address to raft listen")
	fs.StringVar(&cfg.Node.ListenClientAddr, "listen-client-addr", DefaultNodeListenClientAddr, "address to grpc listen")
	fs.Var(newUInt32Value(DefaultNodeMaxSnapshots, &cfg.Node.MaxSnapshots), "max-snapshots", "max number to snapshots(raft)")

	// main config::store
	fs.Var(newValidatorStringValue[EnumStoreBackend](DefaultStoreBackend, &cfg.Store.Backend), "store-backend", "")

	// main config::store::nuts
	fs.Int64Var(&cfg.Store.Nuts.NodeNum, "nuts-node-num", DefaultStoreNutsNodeNum, "nth node in the system")
	fs.BoolVar(&cfg.Store.Nuts.Sync, "nuts-sync", DefaultStoreNutsSync, "enable sync write")
	fs.BoolVar(&cfg.Store.Nuts.StrictMode, "nuts-strict-mode", DefaultStoreNutsStrictMode, "enable strict mode")

	// main config::cluster
	fs.Var(newValidatorStringValue[EnumClusterState](DefaultClusterState, &cfg.Cluster.State), "cluster-state", "status of the cluster at startup")
	fs.StringVar(&cfg.Cluster.Token, "cluster-token", "", "")

	// main config::cluster::bootstrap(s)
	// in cli: node-1@peer_addr,node-2@peer_addr
	fs.Var(newClusterBootstrapsValue("", &cfg.Cluster.Bootstrap), "cluster-bootstrap", "bootstrap at cluster startup, e.g. : node-1@peer_addr,node-2@peer_addr")

	// main config::security
	fs.BoolVar(&cfg.Security.EnableTLS, "enable-tls", false, "")
	fs.BoolVar(&cfg.Security.AutoTLS, "auto-tls", false, "")
	fs.BoolVar(&cfg.Security.EnablePeerTLS, "enable-peer-tls", false, "")
	fs.BoolVar(&cfg.Security.PeerAutoTLS, "peer-auto-tls", false, "")
	fs.StringVar(&cfg.Security.TLSCert, "tls-cert", "", "")
	fs.StringVar(&cfg.Security.TLSKey, "tls-key", "", "")
	fs.StringVar(&cfg.Security.PeerTLSCert, "peer-tls-cert", "", "")
	fs.StringVar(&cfg.Security.PeerTLSKey, "peer-tls-key", "", "")

	// main config::log
	fs.Var(newValidatorStringValue[EnumLogLogger](DefaultLogLogger, &cfg.Log.Logger), "logger", "")
	fs.StringVar(&cfg.Log.OutputDir, "log-dir", DefaultLogOutputDir, "")
	fs.BoolVar(&cfg.Log.Debug, "log-debug", false, "")

	// main config::misc
	fs.BoolVar(&cfg.Misc.PPROF, "d-pprof", false, "")

	// main config::auth
	fs.StringVar(&cfg.Auth.Token, "auth-token", "", "")

	return fs.Parse(args)
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

	defer cfg.setupEnv()

	switch args[0] {
	case "server":
		if err := bindServerFromArgs(cfg, args[1:]...); err != nil {
			return nil, err
		}
	default:
		fmt.Print(usage)
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

	defer cfg.setupEnv()

	if err := bindFromConfigFile(cfg, path); err != nil {
		return nil, err
	}
	return cfg, nil
}

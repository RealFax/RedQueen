package config

import (
	"flag"
	"fmt"
	"github.com/pkg/errors"
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
	NodeNum    int64  `toml:"node-num"`
	Sync       bool   `toml:"sync"`
	StrictMode bool   `toml:"strict-mode"`
	DataDir    string `toml:"data-dir"`
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
	// Logger enum: zap, capnslog
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

func (c Config) Env() ServerEnv {
	return c.env
}

func newConfigEntity() *Config {
	return new(Config)
}

func bindFromArgs(cfg *Config, args ...string) error {
	if len(args) < 1 {
		return errors.New("invalid program args")
	}

	fs := flag.NewFlagSet(args[0], flag.ExitOnError)

	fs.Usage = func() {
		fmt.Fprintln(fs.Output(), "Usage of RedQueen:")
		fs.PrintDefaults()
	}

	fs.StringVar(&cfg.env.configFile, "config-file", "", "config file path")

	// main config::node
	fs.StringVar(&cfg.Node.ID, "node-id", "", "unique node id")
	fs.StringVar(&cfg.Node.DataDir, "data-dir", "./data", "path to the data dir")
	fs.StringVar(&cfg.Node.ListenPeerAddr, "listen-peer-addr", "", "")
	fs.StringVar(&cfg.Node.ListenClientAddr, "listen-client-addr", "", "")
	fs.Var(newUInt32Value(0, &cfg.Node.MaxSnapshots), "max-snapshots", "")

	// main config::store
	fs.Var(newEnumStoreBackendValue("", &cfg.Store.Backend), "store-backend", "")

	// main config::store::nuts
	fs.Int64Var(&cfg.Store.Nuts.NodeNum, "nuts-node-num", 1, "")
	fs.BoolVar(&cfg.Store.Nuts.Sync, "nuts-sync", false, "")
	fs.BoolVar(&cfg.Store.Nuts.StrictMode, "nuts-strict-mode", false, "")
	fs.StringVar(&cfg.Store.Nuts.DataDir, "nuts-data-dir", "", "")

	// main config::cluster

	// main config::cluster::bootstrap(s)

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

	// main config::misc
	fs.BoolVar(&cfg.Misc.PPROF, "d-pprof", false, "")

	// main config::auth
	fs.StringVar(&cfg.Auth.Token, "auth-token", "", "")
	return nil
}

package config

type ServerEnv interface {
	FirstRun() bool
}

type env struct {
	firstRun bool
}

func (r env) FirstRun() bool {
	return r.firstRun
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

func bindFromArgs(cfg *Config) error {

	return nil
}

func newConfigEntity() *Config {
	return new(Config)
}

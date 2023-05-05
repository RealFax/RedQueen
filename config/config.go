package config

type Node struct {
	ID               string `toml:"id"`
	DataDir          string `toml:"data-dir"`
	ListenPeerAddr   string `toml:"listen-peer-addr"`
	ListenClientAddr string `toml:"listen-client-addr"`
	MaxSnapshots     uint32 `toml:"max-snapshots"`
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
	Node     `toml:"node"`
	Cluster  `toml:"cluster"`
	Security `toml:"security"`
	Log      `toml:"log"`
	Misc     `toml:"misc"`
	Auth     `toml:"auth"`
}

func New() {

}

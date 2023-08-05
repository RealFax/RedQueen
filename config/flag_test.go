package config_test

import (
	"testing"

	"github.com/RealFax/RedQueen/config"
)

func TestEncodeClusterBootstraps(t *testing.T) {
	t.Log(config.EncodeClusterBootstraps([]config.ClusterBootstrap{
		{"node1", "10.0.0.2:5290"},
		{"node2", "10.0.0.3:5290"},
		{"node3", "10.0.0.4:5290"},
	}))
}

func TestDecodeClusterBootstraps(t *testing.T) {
	clusters, err := config.DecodeClusterBootstraps("node1@10.0.0.2:5290,node2@10.0.0.3:5290,node3@10.0.0.4:5290")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(clusters)
}

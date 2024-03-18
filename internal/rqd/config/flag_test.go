package config_test

import (
	"github.com/RealFax/RedQueen/internal/rqd/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	bootstraps = []config.ClusterBootstrap{
		{"node1", "10.0.0.2:5290"},
		{"node2", "10.0.0.3:5290"},
		{"node3", "10.0.0.4:5290"},
	}
	bootstrapString = "node1@10.0.0.2:5290,node2@10.0.0.3:5290,node3@10.0.0.4:5290"
)

func TestEncodeClusterBootstraps(t *testing.T) {
	assert.Equal(t, bootstrapString, config.EncodeClusterBootstraps(bootstraps))
}

func TestDecodeClusterBootstraps(t *testing.T) {
	clusters, err := config.DecodeClusterBootstraps("node1@10.0.0.2:5290,node2@10.0.0.3:5290,node3@10.0.0.4:5290")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, clusters, bootstraps)
}

package main

import (
	"context"
	"time"

	"github.com/RealFax/RedQueen/client"
)

type appendCluster struct {
	cfg appendClusterConfig
}

func (c *appendCluster) configEntity() config {
	return &c.cfg
}

func (c *appendCluster) exec() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	call, err := client.New(ctx, endpoints, true)
	if err != nil {
		return err
	}

	return call.AppendCluster(ctx, c.cfg.serverID, c.cfg.peerAddr, c.cfg.voter)
}

func newAppendCluster() *appendCluster {
	return &appendCluster{}
}

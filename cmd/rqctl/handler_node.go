package main

import (
	"github.com/RealFax/RedQueen/client"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
)

func NodeAppendCluster(c *cli.Context) error {
	return invoker.AppendCluster(c.Context, c.String("server-id"), c.String("peer-addr"), true)
}

func NodeLeaderMonitor(c *cli.Context) error {
	_invoker, err := client.New(
		c.Context,
		[]string{c.String("endpoint")},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return err
	}

	receiver := client.NewLeaderMonitorReceiver()
	go func() {
		if err = _invoker.LeaderMonitor(c.Context, receiver); err != nil {
			log.Printf("[-] %s", err)
			close(*receiver)
			return
		}
	}()

	for {
		result, ok := <-*receiver
		if !ok {
			return nil
		}
		log.Printf("[+] state updated, leader: %v", result)
	}
}

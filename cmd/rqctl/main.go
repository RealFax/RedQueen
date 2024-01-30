package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/RealFax/RedQueen/internal/version"
	"github.com/RealFax/RedQueen/pkg/client"
	"net/netip"
	"os"
	"strings"
	"syscall"

	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	helpText = `put the endpoint info into the environment variable: RQ_ENDPOINTS
	format: 172.16.0.100:5230,172.16.0.101:5230,172.16.0.102:5230
`
)

var (
	invoker *client.Client
)

func dialRQ() error {
	endpointString, ok := syscall.Getenv("RQ_ENDPOINTS")
	if !ok {
		fmt.Print(helpText)
		return errors.New("can not read endpoints from environment variable RQ_ENDPOINTS")
	}

	dirtyEndpoints := strings.Split(endpointString, ",")

	var (
		err       error
		endpoints = make([]string, len(dirtyEndpoints))
	)
	for i, endpoint := range dirtyEndpoints {
		addr := strings.TrimSpace(endpoint)
		if _, err = netip.ParseAddrPort(addr); err != nil {
			return err
		}
		endpoints[i] = addr
	}

	client.SetMaxOpenConn(1)
	if invoker, err = client.New(
		context.Background(),
		endpoints,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	); err != nil {
		return err
	}
	return nil
}

func main() {

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Print(version.String())
	}
	app := &cli.App{
		Name:      "rqctl",
		Version:   version.String(),
		Usage:     "cli client for RedQueen 🤖",
		UsageText: helpText,
		Commands: []*cli.Command{
			{
				Name:      "append-cluster",
				UsageText: "Append a new node to the raft cluster",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "server-id",
						Usage: "RedQueen node server id",
					},
					&cli.StringFlag{
						Name:  "peer-addr",
						Usage: "RedQueen peer address",
					},
				},
				Action: NodeAppendCluster,
			}, {
				Name:      "leader-monitor",
				UsageText: "Monitor the election (voting) status of the specified node",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "endpoint",
						Usage: "RedQueen server endpoint address",
					},
				},
				Action: NodeLeaderMonitor,
			}, {
				Name:      "snapshot",
				UsageText: "Snapshot create server snapshot in server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "endpoint",
						Usage: "RedQueen server endpoint address",
					},
					&cli.StringFlag{
						Name:  "path",
						Usage: "The path of the snapshot in the server",
					},
				},
				Action: NodeSnapshot,
			}, {
				Name: "set",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "key",
					},
					&cli.StringFlag{
						Name:  "value",
						Usage: "if you need to set a binary value, please enter it after base64/hex and carry the -base64/-hex flag",
					},
					&cli.StringFlag{
						Name: "namespace",
					},
					&cli.UintFlag{
						Name: "ttl",
					},
					&cli.BoolFlag{
						Name: "hex",
					},
					&cli.BoolFlag{
						Name: "base64",
					},
				},
				Action: RPCSet,
			}, {
				Name: "get",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "key",
					},
					&cli.StringFlag{
						Name: "namespace",
					},
				},
				Action: RPCGet,
			}, {
				Name: "prefix-scan",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "prefix",
					},
					&cli.StringFlag{
						Name:  "reg",
						Usage: "Regular expression scan",
					},
					&cli.StringFlag{
						Name: "namespace",
					},
					&cli.Uint64Flag{
						Name:  "offset",
						Value: 0,
					},
					&cli.Uint64Flag{
						Name:  "limit",
						Value: 10,
					},
				},
				Action: RPCPrefixScan,
			}, {
				Name: "del",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "key",
					},
					&cli.StringFlag{
						Name: "namespace",
					},
				},
				Action: RPCDel,
			}, {
				Name: "watch",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "key",
					},
					&cli.StringFlag{
						Name: "namespace",
					},
					&cli.BoolFlag{
						Name: "ignoreError",
					},
				},
				Action: RPCWatch,
			}, {
				Name: "watch-prefix",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "prefix",
					},
					&cli.StringFlag{
						Name: "namespace",
					},
				},
				Action: RPCWatchPrefix,
			},
		},
	}

	if err := dialRQ(); err != nil {
		fmt.Printf("[-] %s.", err)
		return
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("[-] %s.", err)
		return
	}

	fmt.Println("\n[+] command completed successfully.")
}

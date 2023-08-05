package main

import (
	"flag"
	"fmt"

	"github.com/pkg/errors"
)

const usage = `Usage of RedQueen Util:

	method: append-cluster
	format: ./RedQueenUtil [method] <options>
	case#1: ./RedQueenUtil append-cluster -server-id node-4 -peer-addr 127.0.0.1:2539

`

type config interface {
	in(args ...string) error
}

type appendClusterConfig struct {
	serverID, peerAddr string
	voter              bool
}

func (c *appendClusterConfig) in(args ...string) error {
	if len(args) < 1 {
		return errors.New("invalid append-cluster args")
	}

	flagSet := flag.NewFlagSet("append-cluster", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(flagSet.Output(), "Usage of RedQueen Util:")
		flagSet.PrintDefaults()
	}

	flagSet.StringVar(&c.serverID, "server-id", "", "node id")
	flagSet.StringVar(&c.peerAddr, "peer-addr", "", "node peer addr(raft listen addr)")
	flagSet.BoolVar(&c.voter, "voter", true, "is voter")

	return flagSet.Parse(args)
}

type GenerateMetadataConfig struct {
	Endpoints string `json:"endpoints"`
}

func (c *GenerateMetadataConfig) in(args ...string) error {
	if len(args) < 1 {
		return errors.New("invalid generate-metadata args")
	}

	flagSet := flag.NewFlagSet("generate-metadata", flag.ExitOnError)
	flagSet.Usage = func() {
		fmt.Fprintln(flagSet.Output(), "Usage of RedQueen Util:")
		flagSet.PrintDefaults()
	}

	flagSet.StringVar(&c.Endpoints, "endpoints", "", "nodes addr, format: 127.0.0.1:2539,127.0.0.1:3539,127.0.0.1:4539")

	return flagSet.Parse(args)
}

func readFromArgs(bind config, args ...string) error {
	args = args[1:]
	if len(args) == 0 {
		return errors.New("no subcommand provided")
	}

	switch args[0] {
	case "append-cluster":
		if !enableRemote {
			return errors.New("require exec generate-metadata first")
		}
		return bind.in(args[1:]...)
	case "generate-metadata":
		return bind.in(args[1:]...)
	default:
		return errors.New("unknown subcommand")
	}
}

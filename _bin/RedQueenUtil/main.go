package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const metadataFile string = "./.r_metadata"

var (
	enableRemote bool
	endpoints    []string
)

func init() {
	f, err := os.Open(metadataFile)
	if err != nil {
		return
	}
	defer f.Close()

	var metadata GenerateMetadataConfig
	if err = json.NewDecoder(f).Decode(&metadata); err != nil {
		return
	}

	endpoints = strings.Split(metadata.Endpoints, ",")
	if len(endpoints) == 0 {
		fmt.Println("parse endpoints error")
		os.Exit(0)
	}

	enableRemote = true
}

type handler interface {
	configEntity() config
	exec() error
}

func main() {
	m := map[string]handler{
		"generate-metadata": newGenerateMetadata(),
		"append-cluster":    newAppendCluster(),
	}

	if len(os.Args) < 2 {
		fmt.Print(usage)
		fmt.Println("error: no subcommand provided")
		return
	}

	handle, ok := m[os.Args[1]]
	if !ok {
		fmt.Print(usage)
		fmt.Println("error: unmatched subcommand handler")
		return
	}

	cfg := handle.configEntity()

	if err := readFromArgs(cfg, os.Args...); err != nil {
		fmt.Println("error:", err)
		return
	}

	if err := handle.exec(); err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("success")
}

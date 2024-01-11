# RedQueen

[![Go Report Card](https://goreportcard.com/badge/github.com/RealFax/RedQueen)](https://goreportcard.com/report/github.com/RealFax/RedQueen)
[![CodeQL](https://github.com/RealFax/RedQueen/actions/workflows/codeql.yml/badge.svg)](https://github.com/RealFax/RedQueen/actions/workflows/codeql.yml)
[![build-docker](https://github.com/RealFax/RedQueen/actions/workflows/build-docker.yml/badge.svg)](https://github.com/RealFax/RedQueen/actions/workflows/build-docker.yml)
[![build-release](https://github.com/RealFax/RedQueen/actions/workflows/build-release.yml/badge.svg)](https://github.com/RealFax/RedQueen/actions/workflows/build-release.yml)
[![codecov](https://codecov.io/gh/RealFax/RedQueen/branch/master/graph/badge.svg?token=4JL6XDU245)](https://codecov.io/gh/RealFax/RedQueen)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/RealFax/RedQueen)
[![Releases](https://img.shields.io/github/release/RealFax/RedQueen/all.svg?style=flat-square)](https://github.com/RealFax/RedQueen/releases)
[![LICENSE](https://img.shields.io/github/license/RealFax/RedQueen.svg?style=flat-square)](https://github.com/RealFax/RedQueen/blob/master/LICENSE)

[简体中文](./README_zh.md)

_Inspired by the supercomputer (Red Queen) in "Resident Evil", the distributed key-value store is close to it in the distributed system_

This is a reliable distributed key-value store based on the raft algorithm, and internal provides advanced functions such as distributed-lock...

## Client call
`# go get github.com/RealFax/RedQueen@latest`

[Code example](https://github.com/RealFax/RedQueen/tree/master/client/example)

## Write & Read
_RedQueen based on raft algorithm has the characteristics of single node write (Leader node) and multiple node read (Follower node)._

### Write-only call
- `Set`
- `TrySet`
- `Delete`
- `Lock` <!-- IAF start -->
- `Unlock`
- `TryLock` <!-- IAF end -->

### Read-only call
- `Get`
- `PrefixScan`
- `Watch`

## About Internal Advanced Functions
internal advanced functions require long-term experiments to ensure its reliability

### 🧪 Distributed-lock (experimental functions)
RedQueen internal implements a mutex and provides grpc interface calls

## ⚙️ Parameters
Read order: `Environment Variables | Program Arguments -> Configuration File`

### Environment Variables
- `RQ_CONFIG_FILE <string>` Configuration file path. Note: If this parameter is set, the following parameters will be ignored, and the configuration file will be used.
- `RQ_NODE_ID <string>` Node ID
- `RQ_DATA_DIR <string>` Node data storage directory
- `RQ_LISTEN_PEER_ADDR <string>` Node-to-node communication (Raft RPC) listening address, cannot be `0.0.0.0`
- `RQ_LISTEN_CLIENT_ADDR <string>` Node service listening (gRPC API) address
- `RQ_MAX_SNAPSHOTS <uint32>` Maximum number of snapshots
- `RQ_REQUESTS_MERGED <bool>` Whether to enable request merging
- `RQ_STORE_BACKEND <string [nuts]>` Storage backend (default: nuts)
- `RQ_NUTS_NODE_NUM <int64>`
- `RQ_NUTS_SYNC <bool>` Whether to enable synchronous disk writes
- `RQ_NUTS_STRICT_MODE <bool>` Whether to enable call checking
- `RQ_NUTS_RW_MODE <string [fileio, mmap]>` Write mode
- `RQ_CLUSTER_BOOTSTRAP <string>` Cluster information (e.g., node-1@127.0.0.1:5290, node-2@127.0.0.1:4290)
- `RQ_DEBUG_PPROF <bool>` Enable pprof debugging

### Program Arguments
- `-config-file <string>` Configuration file path. Note: If this parameter is set, the following parameters will be ignored, and the configuration file will be used.
- `-node-id <string>` Node ID
- `-data-dir <string>` Node data storage directory
- `-listen-peer-addr <string>` Node-to-node communication (Raft RPC) listening address, cannot be `0.0.0.0`
- `-listen-client-addr <string>` Node service listening (gRPC API) address
- `-max-snapshots <uint32>` Maximum number of snapshots
- `-requests-merged <bool>` Whether to enable request merging
- `-store-backend <string [nuts]>` Storage backend (default: nuts)
- `-nuts-node-num <int64>`
- `-nuts-sync <bool>` Whether to enable synchronous disk writes
- `-nuts-strict-mode <bool>` Whether to enable call checking
- `-nuts-rw-mode <string [fileio, mmap]>` Write mode
- `-cluster-bootstrap <string>` Cluster information (e.g., node-1@127.0.0.1:5290, node-2@127.0.0.1:4290)
- `-d-pprof <bool>` Enable pprof debugging

### Configuration File
```toml
[node]
id = "node-1"
data-dir = "/tmp/red_queen"
listen-peer-addr = "127.0.0.1:5290"
listen-client-addr = "127.0.0.1:5230"
max-snapshots = 5
requests-merged = false

[store]
# backend options
# nuts
backend = "nuts"
    [store.nuts]
    node-num = 1
    sync = false
    strict-mode = false
    rw-mode = "fileio"

[cluster]
    [[cluster.bootstrap]]
    name = "node-1"
    peer-addr = "127.0.0.1:5290"

    [[cluster.bootstrap]]
    name = "node-2"
    peer-addr = "127.0.0.1:4290"

[misc]
    pprof = false
```

### _About More Usage (e.g., Docker Single/Multi-node Deployment), Please Refer to [**Wiki**](https://github.com/RealFax/RedQueen/wiki)_ 🤩

## 🔍 Third-party
- nutsdb [(Apache License 2.0)](https://github.com/nutsdb/nutsdb/blob/master/LICENSE)
- hashicorp raft [(MPL License 2.0)](https://github.com/hashicorp/raft/blob/main/LICENSE)
- boltdb [(MIT License)](https://github.com/boltdb/bolt/blob/master/LICENSE)
- BurntSushi toml (MIT License)
- google uuid [(BSD-3-Clause License)](https://github.com/google/uuid/blob/master/LICENSE)
- grpc [(Apache License 2.0)](https://github.com/grpc/grpc-go/blob/master/LICENSE)
- protobuf [(BSD-3-Clause License)](https://github.com/protocolbuffers/protobuf-go/blob/master/LICENSE)

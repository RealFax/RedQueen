# RedQueen

[![Go Report Card](https://goreportcard.com/badge/github.com/RealFax/RedQueen)](https://goreportcard.com/report/github.com/RealFax/RedQueen)
[![CodeQL](https://github.com/RealFax/RedQueen/actions/workflows/codeql.yml/badge.svg)](https://github.com/RealFax/RedQueen/actions/workflows/codeql.yml)
[![build-docker](https://github.com/RealFax/RedQueen/actions/workflows/build-docker.yml/badge.svg)](https://github.com/RealFax/RedQueen/actions/workflows/build-docker.yml)
[![build-release](https://github.com/RealFax/RedQueen/actions/workflows/build-release.yml/badge.svg)](https://github.com/RealFax/RedQueen/actions/workflows/build-release.yml)
[![codecov](https://codecov.io/gh/RealFax/RedQueen/branch/master/graph/badge.svg?token=4JL6XDU245)](https://codecov.io/gh/RealFax/RedQueen)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/RealFax/RedQueen)
[![Releases](https://img.shields.io/github/release/RealFax/RedQueen/all.svg?style=flat-square)](https://github.com/RealFax/RedQueen/releases)
[![LICENSE](https://img.shields.io/github/license/RealFax/RedQueen.svg?style=flat-square)](https://github.com/RealFax/RedQueen/blob/master/LICENSE)

[English](./README.md)

_灵感来源于《生化危机》中的超级计算机(Red Queen), 分布式key-value存储在分布式系统中地位与其接近_

这是一个基于raft算法实现的可靠分布式key-value存储, 并在内部提供了诸如 分布式锁...之类的高级功能

## 客户端调用
```
go get github.com/RealFax/RedQueen@latest
```

[代码示例](https://github.com/RealFax/RedQueen/tree/master/client/example)

## 写入 & 读取
_基于raft算法实现的RedQueen具备单节点写入(Leader node)多节点读取(Follower node)的特性_

### 仅写入调用
- `Set`
- `TrySet`
- `Delete`
- `Lock` <!-- IAF start -->
- `Unlock`
- `TryLock` <!-- IAF end -->

### 仅读取调用
- `Get`
- `PrefixScan`
- `Watch`

## 关于内部高级功能
内部高级功能需要进行长时间的实验才能保证他的可靠性

### 🧪 分布式锁 (实验功能)
RedQueen在内部实现了一个互斥锁, 并提供grpc接口调用

## ⚙️ 参数
读取顺序 `环境变量 | 程序参数 -> 配置文件`

### 环境变量
- `RQ_CONFIG_FILE <string>` 配置文件路径. note: 设置该参数后, 将会忽略以下参数, 使用配置文件
- `RQ_NODE_ID <string>`  节点ID
- `RQ_DATA_DIR <string>` 节点数据存储目录
- `RQ_LISTEN_PEER_ADDR <string>` 节点间通信监听(raft rpc)地址, 不可为 `0.0.0.0`
- `RQ_LISTEN_CLIENT_ADDR <string>` 节点服务监听(grpc api)地址
- `RQ_MAX_SNAPSHOTS <uint32>` 最大快照数量
- `RQ_REQUESTS_MERGED <bool>` 是否开启合并请求
- `RQ_STORE_BACKEND <string [nuts]>` 存储后端(默认nuts)
- `RQ_NUTS_NODE_NUM <int64>`
- `RQ_NUTS_SYNC <bool>` 是否启用同步写入磁盘
- `RQ_NUTS_STRICT_MODE <bool>` 是否启用调用检查
- `RQ_NUTS_RW_MODE <string [fileio, mmap]>` 写入模式
- `RQ_CLUSTER_BOOTSTRAP <string>` 集群信息 (例如 node-1@127.0.0.1:5290, node-2@127.0.0.1:4290)
- `RQ_DEBUG_PPROF <bool>` 启用pprof调试

### 程序参数
- `-config-file <string>` 配置文件路径. note: 设置该参数后, 将会忽略以下参数, 使用配置文件
- `-node-id <string>` 节点ID
- `-data-dir <string>` 节点数据存储目录
- `-listen-peer-addr <string>` 节点间通信监听(raft rpc)地址, 不可为 `0.0.0.0`
- `-listen-client-addr <string>` 节点服务监听(grpc api)地址
- `-max-snapshots <uint32>` 最大快照数量
- `-requests-merged <bool>` 是否开启合并请求
- `-store-backend <string [nuts]>` 存储后端(默认nuts)
- `-nuts-node-num <int64>`
- `-nuts-sync <bool>` 是否启用同步写入磁盘
- `-nuts-strict-mode <bool>` 是否启用调用检查
- `-nuts-rw-mode <string [fileio, mmap]>` 写入模式
- `-cluster-bootstrap <string>` 集群信息 (例如 node-1@127.0.0.1:5290, node-2@127.0.0.1:4290)
- `-d-pprof <bool>` 启用pprof调试

### 配置文件
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

### _关于更多用法(例如docker单/多节点部署), 请参考 [**Wiki**](https://github.com/RealFax/RedQueen/wiki)_ 🤩

## 🔍 关键第三方库
- nutsdb [(Apache License 2.0)](https://github.com/nutsdb/nutsdb/blob/master/LICENSE)
- hashicorp raft [(MPL License 2.0)](https://github.com/hashicorp/raft/blob/main/LICENSE)
- boltdb [(MIT License)](https://github.com/boltdb/bolt/blob/master/LICENSE)
- BurntSushi toml (MIT License)
- google uuid [(BSD-3-Clause License)](https://github.com/google/uuid/blob/master/LICENSE)
- grpc [(Apache License 2.0)](https://github.com/grpc/grpc-go/blob/master/LICENSE)
- protobuf [(BSD-3-Clause License)](https://github.com/protocolbuffers/protobuf-go/blob/master/LICENSE)

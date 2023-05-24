# RedQueen

[简体中文](./README_zh.md)

_Inspired by the supercomputer (Red Queen) in "Resident Evil", the distributed key-value store is close to it in the distributed system_

This is a reliable distributed key-value store based on the raft algorithm, and internal provides advanced functions such as distributed-lock, ServiceBridges...

## About Internal Advanced Functions
internal advanced functions require long-term experiments to ensure its reliability

### 🧪 Distributed-lock (experimental functions)
RedQueen internal implements a mutex and provides grpc interface calls

### 🔨 ServiceBridges (unimplemented)
RedQueen internal implement a function similar to Service registration and discovery and provides grpc interface calls

## 🔍 Third-party
- nutsdb [(Apache License 2.0)](https://github.com/nutsdb/nutsdb/blob/master/LICENSE)
- hashicorp raft [(MPL License 2.0)](https://github.com/hashicorp/raft/blob/main/LICENSE)
- boltdb [(MIT License)](https://github.com/boltdb/bolt/blob/master/LICENSE)
- BurntSushi toml (MIT License)
- vmihailenco msgpack [(BSD-2-Clause license)](https://github.com/vmihailenco/msgpack/blob/v5/LICENSE)
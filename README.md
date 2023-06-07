# RedQueen

[![Go Report Card](https://goreportcard.com/badge/github.com/RealFax/RedQueen)](https://goreportcard.com/report/github.com/RealFax/RedQueen)
[![CodeQL](https://github.com/RealFax/RedQueen/actions/workflows/codeql.yml/badge.svg)](https://github.com/RealFax/RedQueen/actions/workflows/codeql.yml)
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

## About Internal Advanced Functions
internal advanced functions require long-term experiments to ensure its reliability

### 🧪 Distributed-lock (experimental functions)
RedQueen internal implements a mutex and provides grpc interface calls

## 🔍 Third-party
- nutsdb [(Apache License 2.0)](https://github.com/nutsdb/nutsdb/blob/master/LICENSE)
- hashicorp raft [(MPL License 2.0)](https://github.com/hashicorp/raft/blob/main/LICENSE)
- boltdb [(MIT License)](https://github.com/boltdb/bolt/blob/master/LICENSE)
- BurntSushi toml (MIT License)
- vmihailenco msgpack [(BSD-2-Clause license)](https://github.com/vmihailenco/msgpack/blob/v5/LICENSE)

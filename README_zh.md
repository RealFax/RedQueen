# RedQueen

[![Go Report Card](https://goreportcard.com/badge/github.com/RealFax/RedQueen)](https://goreportcard.com/report/github.com/RealFax/RedQueen)
[![CodeQL](https://github.com/RealFax/RedQueen/actions/workflows/codeql.yml/badge.svg)](https://github.com/RealFax/RedQueen/actions/workflows/codeql.yml)
[![codecov](https://codecov.io/gh/RealFax/RedQueen/branch/master/graph/badge.svg?token=4JL6XDU245)](https://codecov.io/gh/RealFax/RedQueen)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/RealFax/RedQueen)
[![Releases](https://img.shields.io/github/release/RealFax/RedQueen/all.svg?style=flat-square)](https://github.com/RealFax/RedQueen/releases)
[![LICENSE](https://img.shields.io/github/license/RealFax/RedQueen.svg?style=flat-square)](https://github.com/RealFax/RedQueen/blob/master/LICENSE)

[English](./README.md)

_灵感来源于《生化危机》中的超级计算机(Red Queen), 分布式key-value存储在分布式系统中地位与其接近_

这是一个基于raft算法实现的可靠分布式key-value存储, 并在内部提供了诸如 分布式锁、服务桥...之类的高级功能

## 客户端调用
`# go get github.com/RealFax/RedQueen@latest`

[代码示例](https://github.com/RealFax/RedQueen/tree/master/client/example)

## 关于内部高级功能
内部高级功能需要进行长时间的实验才能保证他的可靠性

### 🧪 分布式锁 (实验功能)
RedQueen在内部实现了一个互斥锁, 并提供grpc接口调用

### 🔨 服务桥 (未完成)
RedQueen在内部实现了类似服务注册与发现的功能, 并提供grpc接口调用

## 🔍 关键第三方库
- nutsdb [(Apache License 2.0)](https://github.com/nutsdb/nutsdb/blob/master/LICENSE)
- hashicorp raft [(MPL License 2.0)](https://github.com/hashicorp/raft/blob/main/LICENSE)
- boltdb [(MIT License)](https://github.com/boltdb/bolt/blob/master/LICENSE)
- BurntSushi toml (MIT License)
- vmihailenco msgpack [(BSD-2-Clause license)](https://github.com/vmihailenco/msgpack/blob/v5/LICENSE)

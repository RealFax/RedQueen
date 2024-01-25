.PHONY: all build
all: build

git_hash = $(shell git rev-parse --short HEAD)
build_ts = $(shell date +%s)

build:
	go build --tags=safety_map -ldflags "-s -w -X 'github.com/RealFax/RedQueen/internal/version.BuildTime=$(build_ts)' -X 'github.com/RealFax/RedQueen/internal/version.BuildVersion=$(git_hash)'" -o ./release/rqd ./cmd/rqd/main.go
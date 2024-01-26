#!/bin/bash

function walk() {
  local dir="$1"
  local walk_dir="$2"

    for f in $(ls "$1"); do

     if [ -f "$dir/$f" ] && [[ $walk_dir != "dir" ]] && [[ "$dir/$f" == *".proto" ]]; then
       echo "building protobuf: $dir/$f"
          protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative "$dir/$f"
    fi

      if [ -d "$dir/$f" ]; then
       if [[ $(ls $dir/$f) == *".proto" ]]; then
           protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $dir/$f/*.proto
      else
        walk "$dir/$f" "dir"
      fi
          echo "building protobuf: $dir/$f"
    fi
  done
}

function main() {
  local api_dir='./api'

  if [ "$1" != "" ]; then
    api_dir=$1
  fi

  walk "$api_dir" "file"
}

main "$@"

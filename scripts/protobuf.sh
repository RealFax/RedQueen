#!/usr/bin/env bash

function walk() {
  local dir="$1"
	local walkDir="$2"

    # shellcheck disable=SC2045
    for f in `ls $1`; do

    		if [ -f "$dir/$f" ] && [[ $walkDir != "dir" ]] && [[ "$dir/$f" == *".proto" ]]; then
    			echo "building protobuf: $dir/$f"
    		    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative "$dir/$f"
    		fi

    		if [ -d "$dir/$f" ]
    		then
    			if [[ `ls $dir/$f` == *".proto" ]]; then
    			    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative $dir/$f/*.proto
    			else
    				walk "$dir/$f" "dir"
    			fi
    		    echo "building protobuf: $dir/$f"
    		fi
    done
}

if [ "$1" == "" ]; then
	$1 = "./api"
    echo "invalid api dir"
else
	walk "$1" "file"
fi
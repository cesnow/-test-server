#!/bin/sh
SRC_DIR=.
DST_DIR=../../..

protoc -I=$SRC_DIR --proto_path=$GOPATH/src:. --go_out=$DST_DIR --go-grpc_out=$DST_DIR $SRC_DIR/*.proto

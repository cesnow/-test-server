#!/bin/sh

SRC_DIR=.
DST_DIR=.

T_PROTO_PATH=../../

protoc -I=$SRC_DIR:$T_PROTO_PATH --proto_path=$GOPATH/src:./ --go_out=$DST_DIR --go-grpc_out=require_unimplemented_servers=false:$DST_DIR $SRC_DIR/*.proto

gofmt -w *.go

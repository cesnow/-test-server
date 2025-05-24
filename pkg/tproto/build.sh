#!/bin/sh

SRC_DIR=.
DST_DIR=.

#protoc -I=$SRC_DIR --proto_path=$GOPATH/src:./ --go_out=$DST_DIR --go-grpc_out=require_unimplemented_servers=false:$DST_DIR $SRC_DIR/*.proto

#gofmt -w *.go


##### ts-proto
#protoc \
#  --plugin=protoc-gen-ts_proto=../../app/web/node_modules/.bin/protoc-gen-ts_proto \
#  --ts_proto_out=$DST_DIR \
#  --ts_proto_opt=outputServices=grpc-js,esModuleInterop=true,forceLong=long \
#  $SRC_DIR/*.proto


###### protobuf-ts/plugin
protoc \
  --plugin=protoc-gen-ts=../../app/web/node_modules/.bin/protoc-gen-ts \
  --ts_out=. \
  $SRC_DIR/*.proto


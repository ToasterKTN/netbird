#!/usr/bin/env bash
set -e

if ! which realpath >/dev/null 2>&1; then
	echo realpath is not installed
	echo run: brew install coreutils
	exit 1
fi

old_pwd=$(pwd)
script_path=$(dirname "$(realpath "$0")")
cd "$script_path"
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.26.3
protoc -I ./ ./daemon.proto \
  --go_out=../ \
  --go-grpc_out=../ \
  --experimental_allow_proto3_optional
# The JSON gateway lives in its own package (client/proto/gateway) so that
# importers of client/proto (e.g. the UI) do not link grpc-gateway.
gw_tmp=$(mktemp -d)
protoc -I ./ ./daemon.proto \
  --grpc-gateway_out="$gw_tmp" \
  --grpc-gateway_opt=generate_unbound_methods=true,standalone=true,Mdaemon.proto=github.com/netbirdio/netbird/client/proto
sed -e 's/^Package proto is a reverse proxy./Package gateway is a reverse proxy./' -e 's/^package proto$/package gateway/' \
  "$gw_tmp/github.com/netbirdio/netbird/client/proto/daemon.pb.gw.go" > ./gateway/daemon.pb.gw.go
rm -rf "$gw_tmp"
cd "$old_pwd"

#!/usr/bin/env bash
set -ex

source scripts/_ldflags.sh
go build -ldflags "${LDFLAGS}" -o $GOPATH/bin/goraffe github.com/spilliams/goraffe/cmd/goraffe

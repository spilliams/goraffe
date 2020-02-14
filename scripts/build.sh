#!/usr/bin/env bash
set -e

o="."
if [[ ! -z "$GOPATH" ]]; then
    o=$GOPATH/bin
fi
source scripts/_ldflags.sh
go build -ldflags "${LDFLAGS}" -o $o/goraffe github.com/spilliams/goraffe/cmd/goraffe
echo "Updated binary $o/goraffe"

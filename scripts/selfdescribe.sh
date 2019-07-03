#!/usr/bin/env bash
set -ex

scripts/build.sh
goraffe imports github.com/spilliams/goraffe cmd/goraffe | dot -Tpng > doc/goraffe.png

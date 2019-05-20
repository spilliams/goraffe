#!/usr/bin/env bash
set -ex

scripts/build.sh
goraffe -v imports github.com/spilliams/goraffe/cli \
    --prefix github.com/spilliams/goraffe/ \
    | dot -Tpng > doc/goraffe.png

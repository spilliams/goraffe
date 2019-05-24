export GITHASH=`git rev-parse --short HEAD`
export BUILDTIME=`date -u '+%Y-%m-%d_%I:%M:%S%p'`
export VERSION="0.2.0-beta"
export LDFLAGS="
    -X github.com/spilliams/goraffe/version.versionNumber=${VERSION}
    -X github.com/spilliams/goraffe/version.gitHash=${GITHASH}
    -X github.com/spilliams/goraffe/version.buildTime=${BUILDTIME}"
export GITHASH=`git rev-parse --short HEAD`
export BUILDTIME=`date -u '+%Y-%m-%d_%I:%M:%S%p'`
export VERSION="0.2.1-beta"
export LDFLAGS="
    -X github.com/spilliams/goraffe/internal/version.versionNumber=${VERSION}
    -X github.com/spilliams/goraffe/internal/version.gitHash=${GITHASH}
    -X github.com/spilliams/goraffe/internal/version.buildTime=${BUILDTIME}"
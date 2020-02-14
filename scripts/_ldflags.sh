export GITHASH=`git rev-parse --short HEAD`
export VERSION=`git describe --tags --contains ${GITHASH} --match v[0-9]*\.[0-9]*\.[0-9]* 2>/dev/null | grep v[0-9]*\.[0-9]*\.[0-9]* -o || echo "development"`
export BUILDTIME=`date -u '+%Y-%m-%d_%I:%M:%S%p'`
export LDFLAGS="
    -X github.com/spilliams/goraffe/internal/version.gitHash=${GITHASH}
    -X github.com/spilliams/goraffe/internal/version.versionNumber=${VERSION}
    -X github.com/spilliams/goraffe/internal/version.buildTime=${BUILDTIME}"

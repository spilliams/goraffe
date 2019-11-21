export GITHASH=`git rev-parse --short HEAD`
export BUILDTIME=`date -u '+%Y-%m-%d_%I:%M:%S%p'`
export LDFLAGS="
    -X github.com/spilliams/goraffe/internal/version.gitHash=${GITHASH}
    -X github.com/spilliams/goraffe/internal/version.buildTime=${BUILDTIME}"
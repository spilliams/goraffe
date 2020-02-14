package importer

import (
	"fmt"
	"go/build"
)

type importError struct {
	parentDirPkg *build.Package
	parentDirErr error
	gopathPkg    *build.Package
	gopathErr    error
	vendorPkg    *build.Package
	vendorErr    error
}

func (ie importError) Error() string {
	return fmt.Sprintf("{parent error: %v; gopath error: %v; vendor error: %v}", ie.parentDirErr, ie.gopathErr, ie.vendorErr)
}

func (ie importError) String() string {
	s := fmt.Sprintf("parent dir package:\n%+v\n", ie.parentDirPkg)
	s += fmt.Sprintf("parent dir error: %v\n", ie.parentDirErr)
	s += fmt.Sprintf("gopath package:\n%+v\n", ie.gopathPkg)
	s += fmt.Sprintf("gopath error; %v\n", ie.gopathErr)
	s += fmt.Sprintf("vendor package:\n%+v\n", ie.vendorPkg)
	s += fmt.Sprintf("vendor error: %v", ie.vendorErr)
	return s
}

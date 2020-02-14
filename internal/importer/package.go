package importer

import (
	"go/build"
	"path"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spilliams/goraffe/pkg/tree"
)

type importedPackage struct {
	root bool
	pkg  *build.Package
	deps []string
}

func NewPackageImporter(parentDirectory string) Importer {
	return &PackageImporter{
		parentDirectory: path.Clean(parentDirectory),
		packageMap:      make(map[string]*importedPackage),
	}
}

type PackageImporter struct {
	parentDirectory string
	packageMap      map[string]*importedPackage
	includeTests    bool
	includeExts     bool
}

// SetIncludeTests modifies the receiver to include or exclude imports from Go
// test files.
func (pi *PackageImporter) SetIncludeTests(includeTests bool) {
	pi.includeTests = includeTests
}

// SetIncludeExts modifies the receiver to include or exclude packages outside
// the receiver's parent directory.
func (pi *PackageImporter) SetIncludeExts(includeExts bool) {
	pi.includeExts = includeExts
}

// Import will attempt to import a package as a "root".
func (pi *PackageImporter) Import(name string) (bool, error) {
	return pi.importRecursive(name, false, true)
}

// ImportRecursive will attempt to import a package as a "root". Then, it
// will recursively call itself on all of that package's imports.
func (pi *PackageImporter) ImportRecursive(name string) (bool, error) {
	return pi.importRecursive(name, true, true)
}

func (pi *PackageImporter) importRecursive(name string, recurse, root bool) (bool, error) {
	// skip the ones we should not include
	if !shouldInclude(name) {
		return false, nil
	}

	// skip the ones we've seen before
	if _, ok := pi.packageMap[name]; ok {
		return false, nil
	}

	pkg, err := pi.importOne(name)
	if err != nil {
		return false, err
	}

	// now that we have a package we can tell if it's external or not
	if !pi.includeExts && !strings.HasPrefix(pkg.ImportPath, pi.parentDirectory) {
		return false, nil
	}

	// save it locally
	node := importedPackage{
		root: root,
		pkg:  pkg,
	}

	// make the list of packages this leaf imports
	deps := pi.filterImports(pkg.Imports)
	if pi.includeTests {
		deps = append(deps, pi.filterImports(pkg.TestImports)...)
	}
	deps = unique(deps)
	sort.Strings(deps)
	node.deps = deps
	pi.packageMap[name] = &node

	// go one level deeper...
	if recurse {
		for _, childPkg := range deps {
			added, err := pi.importRecursive(childPkg, recurse, false)
			if err != nil {
				return added, err
			}
		}
	}

	return true, nil
}

// importOne attempts to import the named package. Note this only works with
// packages in the local GOPATH or in a vendor directory inside the receiver's
// parentDirectory.
func (pi *PackageImporter) importOne(name string) (*build.Package, error) {
	// the given name may or may not have the parent directory prefixed
	name = strings.TrimPrefix(name, pi.parentDirectory)
	parentName := path.Join(pi.parentDirectory, name)
	iErr := importError{}

	// first try the name prefixed with the parent directory
	pPkg, pErr := build.Import(parentName, "", 0)
	if pErr == nil {
		return pPkg, nil
	}
	iErr.parentDirErr = pErr
	iErr.parentDirPkg = pPkg

	// then try the package without a source dir
	gPkg, gErr := build.Import(name, "", 0)
	if gErr == nil {
		return gPkg, nil
	}
	iErr.gopathErr = gErr
	iErr.gopathPkg = gPkg

	// try importing it from vendor...
	name = path.Join("vendor", name)
	vPkg, vErr := build.Import(name, "", 0)
	if vErr == nil {
		return vPkg, nil
	}
	iErr.vendorErr = vErr
	iErr.vendorPkg = vPkg

	return nil, iErr
}

func (pi *PackageImporter) filterImports(imports []string) []string {
	r := []string{}
	for _, name := range imports {
		logrus.Debugf("checking import %s", name)
		if !pi.includeExts && !strings.HasPrefix(name, pi.parentDirectory) {
			logrus.Debugf("  didn't pass ext filter. (%v %v)",
				pi.includeExts, strings.HasPrefix(name, pi.parentDirectory))
			continue
		}
		if !shouldInclude(name) {
			logrus.Debug("  didn't pass include filter")
			continue
		}
		r = append(r, name)
	}
	return r
}

func shouldInclude(name string) bool {
	if name == "C" {
		return false
	}

	return true
}

func unique(st []string) []string {
	var r []string
	for _, s := range st {
		if !contains(r, s) {
			r = append(r, s)
		}
	}
	return r
}

func contains(st []string, s string) bool {
	for _, cmp := range st {
		if s == cmp {
			return true
		}
	}
	return false
}

func (pi *PackageImporter) Tree() *tree.Tree {
	// TODO
	return nil
}

package tree

import (
	"fmt"
	"go/build"
	"path"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// Add will attempt to add a package to the tree as a "root".
func (t *Tree) Add(name string) (bool, error) {
	return t.add(name, false, true)
}

// AddRecursive will attempt to add a package to the tree as a "root". Then, it
// will recursively call itself on all of that package's imports.
func (t *Tree) AddRecursive(name string) (bool, error) {
	return t.add(name, true, true)
}

func (t *Tree) add(name string, recurse, root bool) (bool, error) {
	logrus.Debugf("add(%s, %v, %v)", name, recurse, root)

	// skip the ones we should not include
	if !t.shouldInclude(name) {
		return false, nil
	}

	// skip the ones we've seen before
	if _, ok := t.packageMap[name]; ok {
		return false, nil
	}

	pkg, iErr := t.importPkg(name)
	if iErr != nil {
		// we had trouble importing this, which means it's not a local package
		if !t.includeExts {
			return false, nil
		}
	}

	if pkg == nil {
		// either we should include externals and got an error back,
		// or something went very wrong.
		// either way, add the leaf to the tree. It will be "broken" because it
		// doesn't have a pkg
		leaf := NewLeaf(name)
		leaf.SetRoot(root)
		t.packageMap[name] = leaf
		return true, nil
	}

	// only keep the externals if that flag is true
	if !t.includeExts && !strings.HasPrefix(pkg.ImportPath, t.parentDirectory) {
		return false, nil
	}

	if typedErr, ok := iErr.(importError); ok {
		logrus.Debugf("error from import: %s", typedErr.String())
	}

	logrus.Debugf("adding %s", name)
	leaf := NewLeaf(name)
	leaf.SetRoot(root)

	// we got past the external checks and still have an import error
	if iErr != nil {
		t.packageMap[name] = leaf
		return true, nil
	}

	displayName := strings.TrimPrefix(name, t.parentDirectory)
	displayName = strings.TrimPrefix(displayName, "/")
	leaf.SetDisplayName(displayName)

	// we still want to include broken leaves I think
	leaf.pkg = pkg

	// make the list of packages this leaf imports
	deps := t.filterImports(pkg.Imports)
	if t.includeTests {
		deps = append(deps, t.filterImports(pkg.TestImports)...)
	}
	deps = unique(deps)
	sort.Strings(deps)
	leaf.deps = deps
	t.packageMap[name] = leaf

	// go one level deeper...
	if recurse {
		for _, childPkg := range deps {
			added, err := t.add(childPkg, recurse, false)
			if err != nil {
				return added, err
			}
		}
	}

	return true, nil
}

func (t *Tree) filterImports(imports []string) []string {
	r := []string{}
	for _, name := range imports {
		logrus.Debugf("checking import %s", name)
		if !t.includeExts && !strings.HasPrefix(name, t.parentDirectory) {
			logrus.Debugf("  didn't pass ext filter. (%v %v)",
				t.includeExts, strings.HasPrefix(name, t.parentDirectory))
			continue
		}
		if !t.shouldInclude(name) {
			logrus.Debug("  didn't pass include filter")
			continue
		}
		r = append(r, name)
	}
	return r
}

func (t *Tree) shouldInclude(name string) bool {
	if name == "C" {
		return false
	}

	return true
}

type importError struct {
	parentDirPkg *build.Package
	parentDirErr error
	gopathPkg    *build.Package
	gopathErr    error
	vendorPkg    *build.Package
	vendorErr    error
}

func (i importError) Error() string {
	return fmt.Sprintf("{parent error: %v; gopath error: %v; vendor error: %v}", i.parentDirErr, i.gopathErr, i.vendorErr)
}

func (i importError) String() string {
	s := fmt.Sprintf("parent dir package:\n%+v\n", i.parentDirPkg)
	s += fmt.Sprintf("parent dir error: %v\n", i.parentDirErr)
	s += fmt.Sprintf("gopath package:\n%+v\n", i.gopathPkg)
	s += fmt.Sprintf("gopath error; %v\n", i.gopathErr)
	s += fmt.Sprintf("vendor package:\n%+v\n", i.vendorPkg)
	s += fmt.Sprintf("vendor error: %v", i.vendorErr)
	return s
}

func (t *Tree) importPkg(name string) (*build.Package, error) {
	// the given name may or may not have the parent directory prefixed
	name = strings.TrimPrefix(name, t.parentDirectory)
	parentName := path.Join(t.parentDirectory, name)
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

// Keep marks a single package in the receiver for keeping.
// Should only be used as a result of user action (e.g. not for automatic
// "grow" operations)
func (t *Tree) Keep(name string) error {
	leaf, ok := t.packageMap[name]
	if !ok {
		return fmt.Errorf("package %s not found", name)
	}
	leaf.keep = true
	leaf.userKeep = true
	t.packageMap[name] = leaf
	return nil
}

// Grow expands the "tree" of kept packages by the given count. This works in
// both directions (ancestors and descendants).
func (t *Tree) Grow(count int) {
	t.printGrow(count, "Before")

	if count <= 0 {
		return
	}

	// make a copy of the packagemap
	copy := t.copyPackageMap()

	// only mutate the copy, not the original
	// grow down
	for _, leaf := range t.packageMap {
		if leaf.keep {
			for _, importName := range leaf.deps {
				if _, ok := copy[importName]; ok {
					copy[importName].keep = true
				}
			}
		}
	}

	// grow up
	for name, leaf := range t.packageMap {
		for _, importName := range leaf.deps {
			if upLeaf, ok := t.packageMap[importName]; ok {
				if upLeaf.keep {
					copy[name].keep = true
				}
			}
		}
	}

	t.packageMap = copy
	t.printGrow(count, "After")

	t.Grow(count - 1)
}

func (t *Tree) printGrow(count int, info string) {
	keepCount := 0
	totalCount := 0
	for _, leaf := range t.packageMap {
		totalCount++
		if leaf.keep {
			keepCount++
		}
	}
	logrus.Debugf("Grow %d. %s: keep %d (total %d)", count, info, keepCount, totalCount)
}

func (t *Tree) copyPackageMap() map[string]*Leaf {
	copy := make(map[string]*Leaf)

	for name, leaf := range t.packageMap {
		copy[name] = leaf.copy()
	}

	return copy
}

// Prune removes all packages and package imports from the tree that are not
// marked for keeping.
func (t *Tree) Prune() {
	for name, leaf := range t.packageMap {
		if !leaf.keep {
			delete(t.packageMap, name)
			continue
		}
		newDeps := []string{}
		for _, importName := range leaf.deps {
			if importLeaf, ok := t.packageMap[importName]; ok && importLeaf.keep {
				newDeps = append(newDeps, importName)
			}
			leaf.deps = newDeps
			t.packageMap[name] = leaf
		}
	}
}

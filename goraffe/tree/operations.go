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
	displayName := strings.TrimPrefix(name, t.parentDirectory)
	name = path.Join(t.parentDirectory, displayName)

	// make sure the name passes the filter
	name = t.filter(name)
	if name == "" {
		return false, nil
	}

	// C isn't really a package
	if name == "C" {
		t.packageMap["C"] = nil
	}

	// skip the ones we've seen before
	if _, ok := t.packageMap[name]; ok {
		return false, nil
	}

	// not skipping this package, make a new leaf for it
	logrus.Debugf("adding %s", displayName)
	leaf := NewLeaf(displayName)
	leaf.SetRoot(root)

	// try importing it
	pkg, err := build.Import(name, "", 0)
	if err != nil {
		// try importing it from vendor...
		name = path.Join("vendor", name)
		pkg, err = build.Import(name, "", 0)
	}
	// if we failed to import it, the leaf remains as a "broken" leaf

	// continue with the leaves that we did find
	if err == nil {
		leaf.pkg = pkg

		// make the list of packages this leaf imports
		deps := t.filterNames(pkg.Imports)
		if t.includeTests {
			deps = append(deps, t.filterNames(pkg.TestImports)...)
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
	}

	return true, nil
}

func (t *Tree) filterNames(names []string) []string {
	r := []string{}
	for _, name := range names {
		name = t.filter(name)
		if name != "" {
			r = append(r, name)
		}
	}
	return r
}

func (t *Tree) filter(name string) string {
	if !strings.HasPrefix(name, t.parentDirectory) {
		// doesn't have the prefix, skip it
		return ""
	}

	return name
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

package tree

import (
	"fmt"
	"go/build"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
)

func (t *Tree) Add(name string) (bool, error) {
	return t.add(name, false, true)
}

func (t *Tree) AddRecursive(name string) (bool, error) {
	return t.add(name, true, true)
}

// add attempts to add a package to the tree. This will run the package through
// and applicable filters, and if it passes them, adds it to the tree. Then the
// package's Imports list will also be filtered. If `recurse` is true, add will
// run on each of the package's (filtered) Imports. If `root` is true, the new
// leaf will be marked as a "root" package
func (t *Tree) add(name string, recurse, root bool) (bool, error) {
	name = t.filter(name)
	if name == "" {
		return false, nil
	}

	displayName := strings.TrimPrefix(name, t.prefix)

	if name == "C" {
		// C isn't really a package
		t.packageMap["C"] = nil
	}

	if _, ok := t.packageMap[name]; ok {
		// seen this package before, skip it
		return false, nil
	}

	if strings.HasPrefix(name, "golang_org") {
		displayName = path.Join("vendor", name)
	}

	logrus.Debugf("adding %s", name)

	pkg, err := build.Import(name, "", 0)
	if err != nil {
		return false, err
	}
	pkg.Imports = t.filterNames(pkg.Imports)
	t.packageMap[name] = &Leaf{pkg: pkg, displayName: displayName, root: root}

	if recurse {
		for _, childPkg := range t.packageMap[name].pkg.Imports {
			if added, err := t.add(childPkg, recurse, false); err != nil {
				return added, err
			}
		}
	}

	return true, nil
}

// Keep marks a single package in the tree for keeping.
func (t *Tree) Keep(name string) (err error) {
	leaf, ok := t.packageMap[name]
	if !ok {
		return fmt.Errorf("package %s not found", name)
	}
	leaf.keep = 2
	t.packageMap[name] = leaf
	return nil
}

// Grow expands the "tree" of kept packages by the given count. This works in
// both directions (ancestors and descendants).
func (t *Tree) Grow(count int) {
	keepCount := 0
	totalCount := 0
	for _, leaf := range t.packageMap {
		totalCount++
		if leaf.keep > 0 {
			keepCount++
		}
	}
	logrus.Debugf("Grow %d. Before: keep %d (total %d)", count, keepCount, totalCount)

	if count <= 0 {
		return
	}

	// make a copy of the packagemap
	copy := t.copyPackageMap()

	// only mutate the copy, not the original
	// grow down
	for _, leaf := range t.packageMap {
		if leaf.keep == 2 {
			for _, importName := range leaf.pkg.Imports {
				if _, ok := copy[importName]; ok {
					copy[importName].keep = 1
				}
			}
		}
	}

	// grow up
	for name, leaf := range t.packageMap {
		for _, importName := range leaf.pkg.Imports {
			if upLeaf, ok := t.packageMap[importName]; ok {
				if upLeaf.keep == 2 {
					copy[name].keep = 1
				}
			}
		}
	}

	keepCount = 0
	for _, leaf := range copy {
		if leaf.keep > 0 {
			keepCount++
		}
	}
	logrus.Debugf("Grow %d. After: keep %d", count, keepCount)

	t.packageMap = copy

	t.Grow(count - 1)
}

// Prune removes all packages and package imports from the tree that are not
// marked for keeping.
func (t *Tree) Prune() {
	for name, leaf := range t.packageMap {
		if leaf.keep == 0 {
			delete(t.packageMap, name)
			continue
		}
		newImports := []string{}
		for _, importName := range leaf.pkg.Imports {
			if importLeaf, ok := t.packageMap[importName]; ok && importLeaf.keep > 0 {
				newImports = append(newImports, importName)
			}
			leaf.pkg.Imports = newImports
			t.packageMap[name] = leaf
		}
	}
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
	if t.filterPattern != nil && !t.filterPattern.MatchString(name) {
		// doesn't match the filter, skip it
		return ""
	}

	if t.prefix != "" {
		if !strings.HasPrefix(name, t.prefix) {
			// doesn't have the prefix, skip it
			return ""
		}
	}

	return name
}

func (t *Tree) copyPackageMap() map[string]*Leaf {
	copy := make(map[string]*Leaf)

	for name, leaf := range t.packageMap {
		copy[name] = leaf.copy()
	}

	return copy
}

func (t *Tree) countImports() {
	// reset import counts
	for _, leaf := range t.packageMap {
		leaf.importCount = 0
	}

	// count again
	for _, leaf := range t.packageMap {
		for _, importName := range leaf.pkg.Imports {
			if importLeaf, ok := t.packageMap[importName]; ok {
				importLeaf.importCount++
				t.packageMap[importName] = importLeaf
			}
		}
	}
}

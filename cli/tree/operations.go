package tree

import (
	"fmt"
	"go/build"
	"path"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

func (t *Tree) Add(name string, includeTests bool) (bool, error) {
	return t.add(name, false, true, includeTests)
}

func (t *Tree) AddRecursive(name string, includeTests bool) (bool, error) {
	return t.add(name, true, true, includeTests)
}

// add attempts to add a package to the tree. This will run the package through
// and applicable filters, and if it passes them, adds it to the tree. Then the
// package's Imports list will also be filtered. If `recurse` is true, add will
// run on each of the package's (filtered) Imports. If `root` is true, the new
// leaf will be marked as a "root" package.
func (t *Tree) add(name string, recurse, root, includeTests bool) (bool, error) {
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

	// logrus.Debugf("adding %s", name)
	leaf := NewLeaf(displayName)
	leaf.SetRoot(root)

	pkg, err := build.Import(name, "", 0)
	if err != nil {
		// try vendoring it
		name = path.Join("vendor", name)
		pkg, err = build.Import(name, "", 0)
	}

	t.packageMap[name] = leaf

	// we can only continue with the ones that are not broken
	if err == nil {
		leaf.pkg = pkg

		deps := t.filterNames(pkg.Imports)
		if includeTests {
			deps = append(deps, t.filterNames(pkg.TestImports)...)
		}
		deps = unique(deps)
		sort.Strings(deps)
		leaf.deps = deps
		t.packageMap[name] = leaf

		if recurse {
			for _, childPkg := range t.packageMap[name].deps {
				added, err := t.add(childPkg, recurse, false, includeTests)
				if err != nil {
					return added, err
				}
			}
		}
	}

	return true, nil
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

// Keep marks a single package in the tree for keeping.
func (t *Tree) Keep(name string) (err error) {
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
	keepCount := 0
	totalCount := 0
	for _, leaf := range t.packageMap {
		totalCount++
		if leaf.keep {
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

	keepCount = 0
	for _, leaf := range copy {
		if leaf.keep {
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
		if leaf != nil {
			leaf.importCount = 0
		}
	}

	// count again
	for _, leaf := range t.packageMap {
		if leaf == nil {
			continue
		}
		for _, importName := range leaf.deps {
			importLeaf, ok := t.packageMap[importName]
			if ok && importLeaf != nil {
				importLeaf.importCount++
				t.packageMap[importName] = importLeaf
			}
		}
	}
}

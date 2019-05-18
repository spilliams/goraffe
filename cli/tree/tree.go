package tree

import (
	"encoding/json"
	"fmt"
	"go/build"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/awalterschulze/gographviz"
	"github.com/sirupsen/logrus"
)

// Tree isn't actually a tree structure, but maintains a map of package names
// to packages.
type Tree struct {
	packageMap    map[string]*leaf
	filterPattern *regexp.Regexp
	prefix        string
}

type leaf struct {
	pkg         *build.Package
	displayName string
	keep        bool
}

// NewTree returns a new, empty Tree
func NewTree() *Tree {
	t := Tree{
		packageMap: make(map[string]*leaf),
	}

	return &t
}

// SetFilter sets the tree's filter. Any time a package is about to be added to
// the tree, it gets checked by this filter first. If it doesn't match the
// regular expression, it won't get added.
func (t *Tree) SetFilter(filter string) (err error) {
	if t.filterPattern, err = regexp.Compile(filter); err != nil {
		return err
	}
	return nil
}

// SetPrefix sets the tree's prefix. Any time a package is about to be added to
// the tree, it gets checked for this prefix. If it doesn't have the prefix, it
// won't get added. Additionally, if the prefix is set, all display names of
// the tree's packages will omit the prefix for clarity.
func (t *Tree) SetPrefix(prefix string) {
	t.prefix = prefix
}

// Add attempts to add a package to the tree. This will run the package through
// and applicable filters, and if it passes them, adds it to the tree. Then the
// package's Imports list will also be filtered. Finally, if `recurse` is set,
// Add will run on each of the package's (filtered) Imports.
func (t *Tree) Add(name string, recurse bool) (err error) {
	logrus.Debugf("adding %s", name)
	name = t.filter(name)
	if name == "" {
		return
	}

	displayName := strings.TrimPrefix(name, t.prefix)

	if name == "C" {
		// C isn't really a package
		t.packageMap["C"] = nil
	}

	if _, ok := t.packageMap[name]; ok {
		// seen this package before, skip it
		return nil
	}

	if strings.HasPrefix(name, "golang_org") {
		displayName = path.Join("vendor", name)
	}

	pkg, err := build.Import(name, "", 0)
	if err != nil {
		return err
	}
	pkg.Imports = t.filterNames(pkg.Imports)
	t.packageMap[name] = &leaf{pkg: pkg, displayName: displayName}

	if recurse {
		for _, childPkg := range t.packageMap[name].pkg.Imports {
			if err = t.Add(childPkg, recurse); err != nil {
				return err
			}
		}
	}

	return nil
}

// Keep marks a single package in the tree for keeping.
func (t *Tree) Keep(name string) (err error) {
	return nil
}

// Grow expands the "tree" of kept packages by the given count. This works in
// both directions (ancestors and descendants).
func (t *Tree) Grow(count int) (err error) {
	return nil
}

// Prune removes all packages and package imports from the tree that are not
// marked for keeping.
func (t *Tree) Prune() (err error) {
	return nil
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

// Broaden returns a list of maps. Each map has one key, and that key and value
// each represent a package name. Each map represents "this package imports
// that package".
func (t *Tree) Broaden() []map[string]string {
	b := make([]map[string]string, 0, len(t.packageMap))
	for name, leaf := range t.packageMap {
		for _, importName := range leaf.pkg.Imports {
			m := map[string]string{name: importName}
			b = append(b, m)
		}
	}
	return b
}

// PackageNames returns a sorted list of all of the tree's packages and package
// Imports.
func (t *Tree) PackageNames() []string {
	names := make(map[string]bool)
	for name, leaf := range t.packageMap {
		names[name] = true
		for _, importName := range leaf.pkg.Imports {
			names[importName] = true
		}
	}
	r := make([]string, 0, len(names))
	for name, _ := range names {
		r = append(r, name)
	}
	// maps are unsorted
	sort.Strings(r)
	return r
}

func (t *Tree) String() string {
	treeBytes, err := json.MarshalIndent(t.packageMap, "", "  ")
	if err != nil {
		return err.Error()
	}

	return string(treeBytes)
}

// Graphviz returns the tree's representation in the graphviz source language,
// as for use with the `dot` command-line tool.
// See https://graphviz.org/documentation/ for more information.
func (t *Tree) Graphviz() (string, error) {

	packageNames := t.PackageNames()
	names := make(map[string]string)
	for i, packageName := range packageNames {
		names[packageName] = fmt.Sprintf("N%d", i)
	}

	edges := t.Broaden()

	g := gographviz.NewGraph()
	if err := g.SetDir(true); err != nil {
		return "", err
	}

	for packageName, nodeName := range names {
		displayName := packageName
		if leaf, ok := t.packageMap[packageName]; ok && leaf.displayName != "" {
			displayName = leaf.displayName
		}
		if err := g.AddNode("", nodeName, map[string]string{
			"label": fmt.Sprintf("\"%s\"", displayName),
			"shape": "box",
		}); err != nil {
			return "", err
		}
	}

	for _, edge := range edges {
		for left, right := range edge {
			nodeLeft := names[left]
			nodeRight := names[right]
			if err := g.AddEdge(nodeLeft, nodeRight, true, map[string]string{
				"weight": "1",
			}); err != nil {
				return "", err
			}
		}
	}

	return g.String(), nil
}

package tree

import (
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
// to Leaves.
type Tree struct {
	packageMap    map[string]*Leaf
	filterPattern *regexp.Regexp
	prefix        string
}

// Leaf contains helpful information about each package, like the package
// itself, a friendly display name, and whether or not the tree wants to keep
// it.
type Leaf struct {
	attrs       map[string]string
	displayName string
	keep        bool
	pkg         *build.Package
}

func (l *Leaf) String() string {
	keepString := ""
	if l.keep {
		keepString = ", keep"
	}
	return fmt.Sprintf("Leaf{%s, %d imports%s}", l.displayName, len(l.pkg.Imports), keepString)
}

func (l *Leaf) copy() *Leaf {
	newLeaf := Leaf{
		attrs:       l.attrs,
		displayName: l.displayName,
		keep:        l.keep,
		pkg:         l.pkg,
	}
	return &newLeaf
}

func (l *Leaf) attributes() map[string]string {
	attr := map[string]string{
		"label": fmt.Sprintf("\"%s\"", l.displayName),
		"shape": "box",
	}
	for k, v := range l.attrs {
		attr[k] = v
	}
	return attr
}

func (l *Leaf) addAttribute(key, value string) {
	if l.attrs == nil {
		l.attrs = make(map[string]string)
	}
	l.attrs[key] = value
}

// NewTree returns a new, empty Tree
func NewTree() *Tree {
	t := Tree{
		packageMap: make(map[string]*Leaf),
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
	t.packageMap[name] = &Leaf{pkg: pkg, displayName: displayName}

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
	leaf, ok := t.packageMap[name]
	if !ok {
		return fmt.Errorf("package %s not found", name)
	}
	leaf.keep = true
	leaf.addAttribute("style", "bold")
	leaf.addAttribute("color", "green")
	t.packageMap[name] = leaf
	return nil
}

// Grow expands the "tree" of kept packages by the given count. This works in
// both directions (ancestors and descendants).
func (t *Tree) Grow(count int) (err error) {
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
		return nil
	}

	// make a copy of the packagemap
	copy := t.copyPackageMap()

	// only mutate the copy, not the original
	// grow down
	for _, leaf := range t.packageMap {
		if leaf.keep {
			for _, importName := range leaf.pkg.Imports {
				if _, ok := copy[importName]; ok {
					copy[importName].keep = true
				}
			}
		}
	}

	// grow up
	for name, leaf := range t.packageMap {
		for _, importName := range leaf.pkg.Imports {
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

	return t.Grow(count - 1)
}

// Prune removes all packages and package imports from the tree that are not
// marked for keeping.
func (t *Tree) Prune() {
	for name, leaf := range t.packageMap {
		if !leaf.keep {
			delete(t.packageMap, name)
			continue
		}
		newImports := []string{}
		for _, importName := range leaf.pkg.Imports {
			if importLeaf, ok := t.packageMap[importName]; ok && importLeaf.keep {
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
	r := "Tree{\n"
	for name, leaf := range t.packageMap {
		r += fmt.Sprintf("\t%s: %s\n", name, leaf)
	}
	r += "}\n"
	return r
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
		leaf, ok := t.packageMap[packageName]
		if !ok {
			return "", fmt.Errorf("Unexpected error: couldn't find %s in package map", packageName)
		}
		if err := g.AddNode("", nodeName, leaf.attributes()); err != nil {
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

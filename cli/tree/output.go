package tree

import (
	"fmt"
	"sort"

	"github.com/awalterschulze/gographviz"
)

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

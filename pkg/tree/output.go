package tree

import (
	"fmt"
	"path"
	"sort"

	"github.com/awalterschulze/gographviz"
	"github.com/sirupsen/logrus"
)

// Broaden returns a list of maps. Each map has one key, and that key and value
// each represent a package name. Each map represents "this package imports
// that package".
func (t *Tree) Broaden() []map[string]string {
	b := make([]map[string]string, 0, len(t.packageMap))
	for name, leaf := range t.packageMap {
		if leaf == nil {
			continue
		}
		for _, importName := range leaf.deps {
			m := map[string]string{name: importName}
			b = append(b, m)
		}
	}
	return b
}

// PackageNames returns a sorted list of all of the tree's packages and package
// dependencies.
func (t *Tree) PackageNames() []string {
	names := make(map[string]bool)
	for name, leaf := range t.packageMap {
		if leaf == nil {
			continue
		}
		names[name] = true
		for _, importName := range leaf.deps {
			names[importName] = true
		}
	}
	r := make([]string, 0, len(names))
	for name := range names {
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

// Stats returns a string of the receiver's statistics
func (t *Tree) Stats() string {
	t.countImports()

	singleParentCount := 0
	rootCount := 0
	brokenCount := 0
	for _, leaf := range t.packageMap {
		if leaf.ImportCount() == 1 {
			singleParentCount++
		}
		if leaf.IsRoot() {
			rootCount++
		}
		if leaf.IsBroken() {
			brokenCount++
		}
	}

	edges := t.Broaden()

	return fmt.Sprintf("%d packages\n  %d with a single parent\n  %d are roots\n  %d are broken\n%d import statements",
		len(t.packageMap),
		singleParentCount,
		rootCount,
		brokenCount,
		len(edges),
	)
}

// Graphviz returns the tree's representation in the graphviz source language,
// as for use with the `dot` command-line tool.
// See https://graphviz.org/documentation/ for more information.
func (t *Tree) Graphviz() (string, error) {

	t.countImports()

	packageNames := t.PackageNames()
	names := make(map[string]string)
	for i, packageName := range packageNames {
		names[packageName] = fmt.Sprintf("N%d", i)
	}

	logrus.Debugf("package names: %v", names)

	edges := t.Broaden()

	logrus.Debugf("edges: %v", edges)

	topGraphName := fmt.Sprintf("\"%s\"", t.parentDirectory)

	g := gographviz.NewGraph()
	if err := g.SetName(topGraphName); err != nil {
		return "", err
	}
	if err := g.SetDir(true); err != nil {
		return "", err
	}

	nodesAdded := []string{}

	// add package nodes
	for packageName, nodeName := range names {
		leaf, ok := t.packageMap[packageName]
		if !ok {
			packageName = path.Join("vendor", packageName)
			leaf, ok = t.packageMap[packageName]
			if !ok {
				logrus.Warnf("couldn't find %s in package map", packageName)
				continue
			}
		}
		if leaf == nil {
			continue
		}
		if err := g.AddNode(topGraphName, nodeName, leaf.attributes()); err != nil {
			return "", err
		}
		nodesAdded = append(nodesAdded, nodeName)
	}

	// add import edges
	for _, edge := range edges {
		for left, right := range edge {
			if contains(nodesAdded, names[left]) && contains(nodesAdded, names[right]) {
				nodeLeft := names[left]
				nodeRight := names[right]
				if err := g.AddEdge(nodeLeft, nodeRight, true, map[string]string{
					"weight": "1",
				}); err != nil {
					return "", err
				}
			}
		}
	}

	// add Legend
	// if err := addLegend(g); err != nil {
	// 	return "", err
	// }

	ast, err := g.WriteAst()
	if err != nil {
		return "", err
	}

	return ast.String(), nil
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

type legend struct {
	key       string
	fillcolor string
	doc       string
}

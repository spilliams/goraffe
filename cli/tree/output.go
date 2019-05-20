package tree

import (
	"fmt"
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

	t.countImports()

	packageNames := t.PackageNames()
	names := make(map[string]string)
	for i, packageName := range packageNames {
		names[packageName] = fmt.Sprintf("N%d", i)
	}

	logrus.Debugf("package names: %v", packageNames)

	edges := t.Broaden()

	logrus.Debugf("edges: %v", edges)

	g := gographviz.NewGraph()
	if err := g.SetName(topGraphName); err != nil {
		return "", err
	}
	if err := g.SetDir(true); err != nil {
		return "", err
	}

	// add package nodes
	for packageName, nodeName := range names {
		leaf, ok := t.packageMap[packageName]
		if !ok {
			logrus.Warnf("couldn't find %s in package map", packageName)
			continue
		}
		if leaf == nil {
			continue
		}
		if err := g.AddNode(topGraphName, nodeName, leaf.attributes()); err != nil {
			return "", err
		}
	}

	// add import edges
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

	// add Legend
	if err := addLegend(g); err != nil {
		return "", err
	}

	ast, err := g.WriteAst()
	if err != nil {
		return "", err
	}

	return ast.String(), nil
}

type legend struct {
	key       string
	fillcolor string
	doc       string
}

func addLegend(g *gographviz.Graph) error {
	err := g.AddSubGraph(topGraphName, legendGraphName, map[string]string{
		"label":   "Legend",
		"style":   "solid",
		"rankdir": "LR",
	})
	if err != nil {
		return err
	}
	items := []legend{
		legend{"broken", BrokenColor, "could not import this package's dependencies"},
		legend{"root", RootColor, "root package (per your command-line args)"},
		legend{"singleParent", SingleParentColor, "only imported by 1 other package"},
		legend{"userKeep", UserKeepColor, "'kept' package (per your command-line flags)"},
	}
	for _, item := range items {
		if err := addLegendItem(g, item); err != nil {
			return err
		}
	}

	return nil
}

func addLegendItem(g *gographviz.Graph, l legend) error {
	if err := g.AddNode(legendGraphName, l.key+"Color", map[string]string{
		"label":     "package",
		"shape":     "box",
		"style":     "filled",
		"fillcolor": fmt.Sprintf("\"%s\"", l.fillcolor),
	}); err != nil {
		return err
	}
	if err := g.AddNode(legendGraphName, l.key+"Doc", map[string]string{
		"label": fmt.Sprintf("\"%s\"", l.doc),
		"shape": "plaintext",
	}); err != nil {
		return err
	}
	if err := g.AddEdge(l.key+"Color", l.key+"Doc", true, map[string]string{
		"style": "invis",
	}); err != nil {
		return err
	}

	return nil
}

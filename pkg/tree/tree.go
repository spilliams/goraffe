package tree

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

func NewTree() *Tree {
	return &Tree{
		leafMap: make(map[string]*Leaf),
	}
}

type Tree struct {
	leafMap map[string]*Leaf
}

func (t *Tree) String() string {
	t.countParents()

	r := "Tree{\n"
	for name, leaf := range t.leafMap {
		r += fmt.Sprintf("\t%s: %s\n", name, leaf)
	}
	r += "}\n"
	return r
}

func (t *Tree) resetParentCounts() {
	// reset import counts
	for _, leaf := range t.leafMap {
		if leaf != nil {
			leaf.parentCount = 0
		}
	}
}

func (t *Tree) countParents() {
	t.resetParentCounts()

	// count again
	for _, leaf := range t.leafMap {
		if leaf == nil {
			continue
		}
		for _, childName := range leaf.children {
			child, ok := t.leafMap[childName]
			if ok && child != nil {
				child.parentCount++
				t.leafMap[childName] = child
			}
		}
	}
}

// Keep marks a single leaf in the tree for keeping.
func (t *Tree) Keep(name string) error {
	return t.keep(name, true)
}

func (t *Tree) keep(name string, userDriven bool) error {
	name = strings.TrimSpace(name)
	logrus.Infof("Keeping %s", name)
	leaf, ok := t.leafMap[name]
	if !ok {
		logrus.Debugf("Current leaves: %+v", t.leafMap)
		return fmt.Errorf("leaf %s not found", name)
	}
	leaf.keep = true
	leaf.userKeep = userDriven
	t.leafMap[name] = leaf
	return nil
}

// Grow expands the subtree of kept leaves by the given count. This works in
// both directions (parents and children).
func (t *Tree) Grow(count int) {
	logrus.Infof("Growing %d", count)
	t.printGrow(count, "Before")

	if count <= 0 {
		return
	}

	copy := t.copyLeafMap()

	// only mutate the copy, not the original
	// grow down
	for _, leaf := range t.leafMap {
		if leaf.keep {
			for _, childName := range leaf.children {
				if _, ok := copy[childName]; ok {
					copy[childName].keep = true
				}
			}
		}
	}

	// grow up
	for name, leaf := range t.leafMap {
		for _, childName := range leaf.children {
			if childLeaf, ok := t.leafMap[childName]; ok {
				if childLeaf.keep {
					copy[name].keep = true
				}
			}
		}
	}

	t.leafMap = copy
	t.printGrow(count, "After")

	t.Grow(count - 1)
}

func (t *Tree) printGrow(count int, info string) {
	keepCount := 0
	totalCount := 0
	for _, leaf := range t.leafMap {
		totalCount++
		if leaf.keep {
			keepCount++
		}
	}
	logrus.Debugf("Grow %d. %s: keep %d (total %d)", count, info, keepCount, totalCount)
}

func (t *Tree) copyLeafMap() map[string]*Leaf {
	copy := make(map[string]*Leaf)

	for name, leaf := range t.leafMap {
		copy[name] = leaf.copy()
	}

	return copy
}

// Prune removes all leaves from the tree that are not marked for keeping.
func (t *Tree) Prune() {
	logrus.Info("Pruning")
	for name, leaf := range t.leafMap {
		if !leaf.keep {
			delete(t.leafMap, name)
			continue
		}
		newChildren := []string{}
		for _, childName := range leaf.children {
			if childLeaf, ok := t.leafMap[childName]; ok && childLeaf.keep {
				newChildren = append(newChildren, childName)
			}
			leaf.children = newChildren
			t.leafMap[name] = leaf
		}
	}
}

// Branch marks all packages between the given package and the root for keeping.
// func (t *Tree) Branch(b string) error {
// 	for name, leaf := range t.packageMap {
// 		if leaf.IsRoot() {
// 			if err := t.branchBetween(name, b); err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func (t *Tree) branchBetween(upper, lower string) error {
// 	logrus.Infof("Branching between %s and %s", upper, lower)
// 	if err := t.keep(upper); err != nil {
// 		return err
// 	}
// 	if err := t.keep(lower); err != nil {
// 		return err
// 	}

// 	broad := t.Broaden()
// 	inverse := map[string][]string{}
// 	for _, m := range broad {
// 		for k, v := range m {
// 			if _, ok := inverse[v]; !ok {
// 				inverse[v] = []string{}
// 			}
// 			inverse[v] = append(inverse[v], k)
// 		}
// 	}

// 	check := []string{lower}

// 	for len(check) > 0 {
// 		this := check[0]
// 		check = check[1:]
// 		for _, importer := range inverse[this] {
// 			l := t.packageMap[importer]
// 			l.keep = true
// 			t.packageMap[importer] = l
// 			check = append(check, importer)
// 		}
// 	}

// 	return nil
// }

// Broaden returns a list of maps. Each map has one key, and that key and value
// each represent a package name. Each map represents "this package imports
// that package".
// func (t *Tree) Broaden() []map[string]string {
// 	b := make([]map[string]string, 0, len(t.packageMap))
// 	for name, leaf := range t.packageMap {
// 		if leaf == nil {
// 			continue
// 		}
// 		for _, importName := range leaf.deps {
// 			m := map[string]string{name: importName}
// 			b = append(b, m)
// 		}
// 	}
// 	return b
// }

// PackageNames returns a sorted list of all of the tree's packages
// func (t *Tree) PackageNames() []string {
// 	names := make(map[string]bool)
// 	for name, leaf := range t.packageMap {
// 		if leaf == nil {
// 			continue
// 		}
// 		names[name] = true
// 		for _, importName := range leaf.deps {
// 			names[importName] = true
// 		}
// 	}
// 	r := make([]string, 0, len(names))
// 	for name := range names {
// 		r = append(r, name)
// 	}
// 	// maps are unsorted
// 	sort.Strings(r)
// 	return r
// }

// Stats returns a string of the receiver's statistics
// func (t *Tree) Stats() string {
// 	t.countParents()

// 	singleParentCount := 0
// 	rootCount := 0
// 	brokenCount := 0
// 	for _, leaf := range t.packageMap {
// 		if leaf.ImportCount() == 1 {
// 			singleParentCount++
// 		}
// 		if leaf.IsRoot() {
// 			rootCount++
// 		}
// 		if leaf.IsBroken() {
// 			brokenCount++
// 		}
// 	}

// 	edges := t.Broaden()

// 	return fmt.Sprintf("%d packages\n  %d with a single parent\n  %d are roots\n  %d are broken\n%d import statements",
// 		len(t.packageMap),
// 		singleParentCount,
// 		rootCount,
// 		brokenCount,
// 		len(edges),
// 	)
// }

// Graphviz returns the tree's representation in the graphviz source language,
// as for use with the `dot` command-line tool.
// See https://graphviz.org/documentation/ for more information.
// func (t *Tree) Graphviz() (string, error) {

// 	t.countParents()

// 	packageNames := t.PackageNames()
// 	names := make(map[string]string)
// 	for i, packageName := range packageNames {
// 		names[packageName] = fmt.Sprintf("N%d", i)
// 	}

// 	logrus.Debugf("package names: %v", names)

// 	edges := t.Broaden()

// 	logrus.Debugf("edges: %v", edges)

// 	topGraphName := "root"

// 	g := gographviz.NewGraph()
// 	if err := g.SetName(topGraphName); err != nil {
// 		return "", err
// 	}
// 	if err := g.SetDir(true); err != nil {
// 		return "", err
// 	}

// 	nodesAdded := []string{}

// 	// add package nodes
// 	for packageName, nodeName := range names {
// 		leaf, ok := t.packageMap[packageName]
// 		if !ok {
// 			packageName = path.Join("vendor", packageName)
// 			leaf, ok = t.packageMap[packageName]
// 			if !ok {
// 				logrus.Warnf("couldn't find %s in package map", packageName)
// 				continue
// 			}
// 		}
// 		if leaf == nil {
// 			continue
// 		}
// 		if err := g.AddNode(topGraphName, nodeName, leaf.attributes()); err != nil {
// 			return "", err
// 		}
// 		nodesAdded = append(nodesAdded, nodeName)
// 	}

// 	// add import edges
// 	for _, edge := range edges {
// 		for left, right := range edge {
// 			if contains(nodesAdded, names[left]) && contains(nodesAdded, names[right]) {
// 				nodeLeft := names[left]
// 				nodeRight := names[right]
// 				if err := g.AddEdge(nodeLeft, nodeRight, true, map[string]string{
// 					"weight": "1",
// 				}); err != nil {
// 					return "", err
// 				}
// 			}
// 		}
// 	}

// 	// add Legend
// 	// if err := addLegend(g); err != nil {
// 	// 	return "", err
// 	// }

// 	ast, err := g.WriteAst()
// 	if err != nil {
// 		return "", err
// 	}

// 	return ast.String(), nil
// }

// func contains(st []string, s string) bool {
// 	for _, cmp := range st {
// 		if s == cmp {
// 			return true
// 		}
// 	}
// 	return false
// }

// type legend struct {
// 	key       string
// 	fillcolor string
// 	doc       string
// }

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
)

type Tree struct {
	pattern    *regexp.Regexp
	packageMap map[string][]string
}

func NewTree(filter string) (*Tree, error) {
	treeFilter, err := regexp.Compile(filter)
	if err != nil {
		return nil, err
	}

	t := Tree{
		pattern:    treeFilter,
		packageMap: make(map[string][]string),
	}

	return &t, nil
}

func (t *Tree) String() string {
	treeBytes, err := json.MarshalIndent(t.packageMap, "", "  ")
	if err != nil {
		return err.Error()
	}

	return string(treeBytes)
}

func (t *Tree) FindImport(pkg string) error {
	if !t.pattern.MatchString(pkg) {
		// doesn't match the filter, skip it
		return nil
	}

	if pkg == "C" {
		// C isn't really a package
		t.packageMap["C"] = nil
	}

	if _, ok := t.packageMap[pkg]; ok {
		// seen this package before, skip it
		return nil
	}

	if strings.HasPrefix(pkg, "golang_org") {
		pkg = path.Join("vendor", pkg)
	}

	gopkg, err := build.Import(pkg, "", 0)
	if err != nil {
		return err
	}
	t.packageMap[pkg] = t.filter(gopkg.Imports)

	// recurse
	for _, childPkg := range t.packageMap[pkg] {
		if err = t.FindImport(childPkg); err != nil {
			return err
		}
	}

	return nil
}

func (t *Tree) filter(names []string) []string {
	var filtered []string
	for _, name := range names {
		if t.pattern.MatchString(name) {
			filtered = append(filtered, name)
		}
	}
	return filtered
}

func (t *Tree) PackageMap() map[string][]string {
	return t.packageMap
}

func (t *Tree) Broaden() []map[string]string {
	b := make([]map[string]string, 0, len(t.packageMap))
	for name, imports := range t.packageMap {
		for _, importName := range imports {
			m := map[string]string{name: importName}
			b = append(b, m)
		}
	}
	return b
}

func (t *Tree) PackageNames() []string {
	names := make(map[string]bool)
	for name, imports := range t.packageMap {
		names[name] = true
		for _, importName := range imports {
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
		if err := g.AddNode("", nodeName, map[string]string{
			"label": fmt.Sprintf("\"%s\"", packageName),
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

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

type Tree struct {
	packageMap    map[string]*build.Package // a map of package names to imported packages
	nameMap       map[string]string         // a map of package names to display names
	filterPattern *regexp.Regexp
	prefix        string
}

func NewTree() *Tree {
	t := Tree{
		packageMap: make(map[string]*build.Package),
		nameMap:    make(map[string]string),
	}

	return &t
}

func (t *Tree) SetFilter(filter string) (err error) {
	if t.filterPattern, err = regexp.Compile(filter); err != nil {
		return err
	}
	return nil
}

func (t *Tree) SetPrefix(prefix string) {
	t.prefix = prefix
}

func (t *Tree) String() string {
	treeBytes, err := json.MarshalIndent(t.packageMap, "", "  ")
	if err != nil {
		return err.Error()
	}

	return string(treeBytes)
}

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
	t.packageMap[name] = pkg
	t.nameMap[name] = displayName

	if recurse {
		for _, childPkg := range t.packageMap[name].Imports {
			if err = t.Add(childPkg, recurse); err != nil {
				return err
			}
		}
	}

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

// Broaden returns a list of maps. Each map has one key, and the key is a
// package that imports the value.
func (t *Tree) Broaden() []map[string]string {
	b := make([]map[string]string, 0, len(t.packageMap))
	for name, pkg := range t.packageMap {
		for _, importName := range pkg.Imports {
			m := map[string]string{name: importName}
			b = append(b, m)
		}
	}
	return b
}

func (t *Tree) PackageNames() []string {
	names := make(map[string]bool)
	for name, pkg := range t.packageMap {
		names[name] = true
		for _, importName := range pkg.Imports {
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
		displayName := packageName
		if dn, ok := t.nameMap[packageName]; ok && dn != "" {
			displayName = dn
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

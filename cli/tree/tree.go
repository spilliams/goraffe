package tree

import (
	"fmt"
	"go/build"
	"regexp"
)

// Tree isn't actually a tree structure, but maintains a map of package names
// to Leaves.
type Tree struct {
	packageMap    map[string]*Leaf
	filterPattern *regexp.Regexp
	prefix        string
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

// Leaf contains helpful information about each package, like the package
// itself, a friendly display name, and whether or not the tree wants to keep
// it.
type Leaf struct {
	attrs       map[string]string
	displayName string
	importCount int // the count of packages that import this one
	keep        bool
	pkg         *build.Package
	root        bool // whether this is one of the named root packages
	userKeep    bool
}

func (l *Leaf) copy() *Leaf {
	newLeaf := Leaf{
		attrs:       l.attrs,
		displayName: l.displayName,
		importCount: l.importCount,
		keep:        l.keep,
		pkg:         l.pkg,
		root:        l.root,
		userKeep:    l.userKeep,
	}
	return &newLeaf
}

func (l *Leaf) String() string {
	keepString := ""
	if l.userKeep {
		keepString = ", user-keep"
	} else if l.keep {
		keepString = ", keep"
	}
	rootString := ""
	if l.root {
		rootString = ", root"
	}
	return fmt.Sprintf("Leaf{%s, %d down, %d up%s%s}",
		l.displayName,
		len(l.pkg.Imports),
		l.importCount,
		keepString,
		rootString)
}

func (l *Leaf) attributes() map[string]string {
	attr := map[string]string{
		"label":     fmt.Sprintf("\"%s\n%d up %d down\"", l.displayName, l.importCount, len(l.pkg.Imports)),
		"shape":     "box",
		"style":     "striped",
		"fillcolor": l.fillColor(),
	}

	for k, v := range l.attrs {
		attr[k] = v
	}
	return attr
}

func (l *Leaf) fillColor() string {
	fc := ""
	if l.userKeep {
		fc = "#76E1FE"
	}
	if l.root {
		if fc != "" {
			fc += ":"
		}
		fc += "green"
	}
	if l.importCount == 1 {
		if fc != "" {
			fc += ":"
		}
		fc += "#fcd92d"
	}
	return fmt.Sprintf("\"%s\"", fc)
}

package tree

import (
	"fmt"
	"go/build"
)

// Leaf contains helpful information about each package, like the package
// itself, a friendly display name, and whether or not the tree wants to keep
// it.
type Leaf struct {
	attrs       map[string]string
	deps        []string
	displayName string
	importCount int // the count of packages that import this one
	keep        bool
	pkg         *build.Package
	root        bool // whether this is one of the named root packages
	userKeep    bool
}

// NewLeaf returns a new leaf.
func NewLeaf(displayName string) *Leaf {
	l := Leaf{
		displayName: displayName,
		importCount: 0,
		keep:        false,
		pkg:         nil,
		root:        false,
		userKeep:    false,
	}
	return &l
}

func (l *Leaf) copy() *Leaf {
	newLeaf := Leaf{
		attrs:       l.attrs,
		deps:        l.deps,
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
	brokenString := ""
	if l.IsBroken() {
		brokenString = ", broken"
	}
	return fmt.Sprintf("Leaf{%s, %d down, %d up%s%s%s}",
		l.displayName,
		len(l.deps),
		l.importCount,
		keepString,
		rootString,
		brokenString,
	)
}

// ImportCount returns the receiver's import-count
func (l *Leaf) ImportCount() int {
	return l.importCount
}

// SetDisplayName modifies the receiver's display name
func (l *Leaf) SetDisplayName(name string) {
	l.displayName = name
}

// SetRoot sets whether the receiver is a root or not.
func (l *Leaf) SetRoot(root bool) {
	l.root = root
}

// IsRoot returns whether the receiver is a "root" leaf or not
func (l *Leaf) IsRoot() bool {
	return l.root
}

func (l *Leaf) attributes() map[string]string {
	attr := map[string]string{
		"label":     fmt.Sprintf("\"%s\\n%d up %d down\"", l.displayName, l.importCount, len(l.deps)),
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
		fc = UserKeepColor
	}
	if l.root {
		fc = addColor(fc, RootColor)
	}
	if l.importCount == 1 {
		fc = addColor(fc, SingleParentColor)
	}
	if l.IsBroken() {
		fc = addColor(fc, BrokenColor)
	}
	return fmt.Sprintf("\"%s\"", fc)
}

// IsBroken returns if the receiver is broken or not
func (l *Leaf) IsBroken() bool {
	return l.pkg == nil
}

func addColor(colors, color string) string {
	if colors != "" {
		colors += ":"
	}
	colors += color
	return colors
}

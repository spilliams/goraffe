package tree

import (
	"fmt"
)

type Leaf struct {
	parentCount int // the count of leaves that descend to this one
	children    []string
	displayName string
	keep        bool
	root        bool // whether this is one of the named root packages
	userKeep    bool
}

func NewLeaf(displayName string) *Leaf {
	l := Leaf{
		parentCount: 0,
		children:    make([]string, 0),
		displayName: displayName,
		keep:        false,
		root:        false,
		userKeep:    false,
	}
	return &l
}

func (l *Leaf) copy() *Leaf {
	newLeaf := Leaf{
		parentCount: l.parentCount,
		children:    l.children,
		displayName: l.displayName,
		keep:        l.keep,
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
		len(l.children),
		l.parentCount,
		keepString,
		rootString,
	)
}

// ParentCount returns the receiver's count of its parents
func (l *Leaf) ParentCount() int {
	return l.parentCount
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

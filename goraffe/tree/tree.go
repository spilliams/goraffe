package tree

import (
	"path"

	"github.com/sirupsen/logrus"
)

// Tree isn't actually a tree structure, but maintains a map of package names
// to Leaves.
type Tree struct {
	packageMap      map[string]*Leaf
	parentDirectory string
	includeTests    bool
	includeExts     bool
}

// NewTree returns a new, empty Tree
func NewTree(parentDirectory string) *Tree {
	t := Tree{
		packageMap:      make(map[string]*Leaf),
		parentDirectory: path.Clean(parentDirectory),
	}

	return &t
}

// SetIncludeTests modifies the receiver to include or exclude imports from Go
// test files.
func (t *Tree) SetIncludeTests(includeTests bool) {
	logrus.Debugf("tree include tests? %v", includeTests)
	t.includeTests = includeTests
}

// SetIncludeExts modifies the receiver to include or exclude packages outside
// the receiver's parent directory.
func (t *Tree) SetIncludeExts(includeExts bool) {
	logrus.Debugf("tree include exts? %v", includeExts)
	t.includeExts = includeExts
}

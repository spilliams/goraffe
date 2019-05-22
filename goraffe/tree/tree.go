package tree

// Tree isn't actually a tree structure, but maintains a map of package names
// to Leaves.
type Tree struct {
	packageMap      map[string]*Leaf
	parentDirectory string
	includeTests    bool
}

// NewTree returns a new, empty Tree
func NewTree(parentDirectory string) *Tree {
	t := Tree{
		packageMap:      make(map[string]*Leaf),
		parentDirectory: parentDirectory,
	}

	return &t
}

// SetIncludeTests modifies the receiver to include or exclude imports from Go
// test files.
func (t *Tree) SetIncludeTests(includeTests bool) {
	t.includeTests = includeTests
}

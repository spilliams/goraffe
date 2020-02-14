package importer

import "github.com/spilliams/goraffe/pkg/tree"

type Importer interface {
	SetIncludeTests(bool)
	SetIncludeExts(bool)
	Import(string) (bool, error)
	ImportRecursive(string) (bool, error)
	Tree() *tree.Tree
}

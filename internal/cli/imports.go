package cli

import (
	"fmt"

	"github.com/spilliams/goraffe/pkg/tree"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// the names of the flags
const (
	growFlag   = "grow"
	keepFlag   = "keep"
	testsFlag  = "tests"
	extsFlag   = "exts"
	branchFlag = "branch"
)

var importsFlags struct {
	grow     int
	keeps    []string
	tests    bool
	exts     bool
	branches []string
}

func newImportsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "imports <parent directory> <root packages>",
		Args:    validateImportsArgs,
		Example: "goraffe -v imports github.com/spilliams/goraffe goraffe",
		Short:   "Visualize package imports",
		Long: `Visualize package imports.
	
The parent directory you provide will be treated as a boundary--packages from
outside that directory will not be included by default. Also, in the resulting
output, the name of that parent will be trimmed from the prefix of all the
package names. The root packages can be named with or without the parent
directory prefix.

The root packages you list as arguments to this command form the start of the
import-dependency tree. How the tree develops is determined by the other flags
you provide this command. By default, the roots' dependencies are added
recursively. The output will include the entire dependency chain of the roots,
bounded by the parent directory.

This command outputs DOT language, to be used with a graphviz tool such as
` + "`dot`" + `. For more information, see https://graphviz.org/.
An example of using the output:

goraffe imports github.com/spilliams/goraffe goraffe | dot -Tsvg > graph.svg

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// importTree is a map of "name" -> ["import", "import", ...]
			importTree := tree.NewTree(args[0])

			importTree.SetIncludeTests(importsFlags.tests)
			importTree.SetIncludeExts(importsFlags.exts)

			packages := args[1:]
			for _, pkg := range packages {
				if _, err := importTree.AddRecursive(pkg); err != nil {
					return err
				}
			}

			for _, name := range importsFlags.keeps {
				if err := importTree.Keep(name); err != nil {
					return err
				}
			}

			// honor either keeps or branch, not both
			if len(importsFlags.keeps) > 0 {
				importTree.Grow(importsFlags.grow)
				importTree.Prune()
			} else {
				for _, branch := range importsFlags.branches {
					if err := importTree.Branch(branch); err != nil {
						return err
					}
				}
				if len(importsFlags.branches) > 0 {
					importTree.Prune()
				}
			}

			logrus.Debug(importTree)

			graph, err := importTree.Graphviz()
			if err != nil {
				return err
			}

			fmt.Println(graph)

			logrus.Info(importTree.Stats())

			return nil
		},
	}

	cmd.Flags().IntVar(&importsFlags.grow, growFlag, 1, "How far to \"grow\" the tree away from any kept\npackages. Use with --"+keepFlag+".")
	cmd.Flags().BoolVar(&importsFlags.tests, testsFlag, false, "Whether to include imports from Go test files.")
	cmd.Flags().StringArrayVar(&importsFlags.keeps, keepFlag, []string{}, "Designate some packages to \"keep\", and prune away\nthe rest.")
	cmd.Flags().BoolVar(&importsFlags.exts, extsFlag, false, "[SLOW] Whether to include packages from outside the\nparent directory.")
	cmd.Flags().StringArrayVar(&importsFlags.branches, branchFlag, []string{}, "Designate a package to branch to--the tree will include the root and this branch, and just the imports in between.")

	return cmd
}

func validateImportsArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("must provide at lease two arguments, the parent directory and at least one package (to be the root of the graph)")
	}
	return nil
}

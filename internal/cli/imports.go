package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spilliams/goraffe/internal/importer"
)

// the names of the flags
const (
	growFlag   = "grow"
	testsFlag  = "tests"
	extsFlag   = "exts"
	branchFlag = "branch"
)

var importsFlags struct {
	grow     int
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

goraffe imports github.com/spilliams/goraffe cmd/goraffe | dot -Tsvg > graph.svg

--branch pkg1[,pkg2,...]
    Default: empty
    Limit the tree output to include only the ancestry between the root
    packages and the named branches (inclusive).

--exts
    Default: false
    Whether or not to include package dependencies from outside the given
    module.
    Warning: this may make the tool very slow!

--grow N
    Default: 0
    For use with the --branch flag.
    How much to "grow" the tree beyond the initially-selected packages.

--tests
    Default: false
    Whether or not to inspect (and follow) the packages imported by test files
    (subject to other behavior limits, such as --exts).

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			importer := importer.NewPackageImporter(args[0])
			importer.SetIncludeTests(importsFlags.tests)
			importer.SetIncludeExts(importsFlags.exts)

			packages := args[1:]
			for _, pkg := range packages {
				if _, err := importer.ImportRecursive(pkg); err != nil {
					return err
				}
			}

			t := importer.Tree()
			fmt.Println(t)

			// for _, branch := range importsFlags.branches {
			// 	if err := importTree.Branch(branch); err != nil {
			// 		return err
			// 	}
			// }
			// if len(importsFlags.branches) > 0 {
			// 	importTree.Grow(importsFlags.grow)
			// 	importTree.Prune()
			// }

			// grapher := grapher.New()
			// if err := grapher.SetFormat(grapher.Graphviz); err != nil {
			// 	return err
			// }
			// grapher.AddDigraph(t)
			// graph, err := grapher.Render()
			// if err != nil {
			// 	return err
			// }

			// fmt.Println(string(graph))

			return nil
		},
	}

	cmd.Flags().IntVar(&importsFlags.grow, growFlag, 1, "How far to \"grow\" the tree away from any kept\npackages. Use with --"+branchFlag+".")
	cmd.Flags().BoolVar(&importsFlags.tests, testsFlag, false, "Whether to include imports from Go test files.")
	cmd.Flags().BoolVar(&importsFlags.exts, extsFlag, false, "[SLOW] Whether to include packages from outside the\nparent directory.")
	cmd.Flags().StringSliceVar(&importsFlags.branches, branchFlag, []string{}, "Designate a package to branch to--the tree will include the root and this branch, and just the imports in between.")

	return cmd
}

func validateImportsArgs(cmd *cobra.Command, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("must provide at lease two arguments, the parent directory and at least one package (to be the root of the graph)")
	}
	return nil
}

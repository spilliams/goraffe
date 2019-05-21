package imports

import (
	"fmt"

	"github.com/spilliams/goraffe/cli/tree"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// the names of the flags
const (
	externalsFlag = "includeExternals"
	filterFlag    = "filter"
	growFlag      = "grow"
	keepFlag      = "keep"
	testsFlag     = "includeTests"
)

var importsFlags struct {
	externals bool
	filter    string
	grow      int
	keeps     []string
	tests     bool
}

func init() {
	Cmd.Flags().StringVar(&importsFlags.filter, filterFlag, "", "Regular expression filter to apply to the package\nlist. The filter is applied via regular expressions,\nand operates on the full import path of each package.")
	Cmd.Flags().IntVar(&importsFlags.grow, growFlag, 1, "How far to \"grow\" the tree away from any kept\npackages. Use with --"+keepFlag+".")
	Cmd.Flags().BoolVar(&importsFlags.externals, externalsFlag, false, "[EXPERIMENTAL] Whether to include packages outside\nthe given parent directroy (e.g. golang builtins, or\nvendored packages).")
	Cmd.Flags().BoolVar(&importsFlags.tests, testsFlag, false, "Whether to include imports from Go test files.")
	Cmd.Flags().StringArrayVar(&importsFlags.keeps, keepFlag, []string{}, "Designate some packages to \"keep\", and prune away\nthe rest.")
}

var Cmd = &cobra.Command{
	Use:   "imports <parent directory> <root packages>",
	Args:  validateImportsArgs,
	Short: "Visualize package imports",
	Long: `Visualize package imports.
	
The parent directory you provide will be treated as a boundary--packages from
outside that directory will not be included by default. Also, in the resulting
output, the name of that parent will be trimmed from the prefix of all the
package names. The root packages can be named with or without the parent
directory prefix.

The root packages you list as arguments to this command form the start of the
import-dependency tree. How the tree develops and is output is determined by
the other flags you provide this command.

By default, the output will include everything the roots import, recursively.
The graph will include the entire dependency chain of the packages you list.

`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// importTree is a map of "name" -> ["import", "import", ...]
		importTree := tree.NewTree(args[0])

		packages := args[1:]
		for _, pkg := range packages {
			if _, err := importTree.AddRecursive(pkg); err != nil {
				return err
			}
		}

		if importsFlags.filter != "" {
			if err := importTree.SetFilter(importsFlags.filter); err != nil {
				return err
			}
		}

		if importsFlags.externals {
			importTree.IncludeExternals(true)
		}

		if importsFlags.tests {
			importTree.IncludeTests(true)
		}

		for _, name := range importsFlags.keeps {
			if err := importTree.Keep(name); err != nil {
				return err
			}
		}
		if len(importsFlags.keeps) > 0 {
			importTree.Grow(importsFlags.grow)
			importTree.Prune()
		}

		logrus.Debug(importTree)

		graph, err := importTree.Graphviz()
		if err != nil {
			return err
		}

		fmt.Println(graph)

		return nil
	},
}

func validateImportsArgs(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("must provide at lease one argument, as package root")
	}
	return nil
}

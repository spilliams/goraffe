package imports

import (
	"fmt"

	"github.com/spilliams/goraffe/cli/tree"

	"github.com/spf13/cobra"
)

const (
	pkgFilterFlag = "filter"
	pkgPrefixFlag = "prefix"
)

var importsFlags struct {
	pkgFilter string
	pkgPrefix string
}

var Cmd = &cobra.Command{
	Use:   "imports <packages>",
	Short: "Visualize package imports",
	Long: `Visualize package imports.
	
The packages you list as arguments to this command all get added as "roots" in
the graph.

The graph will include everything the roots import, recursively. The graph will
include the entire dependency chain of the packages you list.

If you provide the optional --` + pkgFilterFlag + ` flag, the graph will not
include packages that don't match that filter.

If you provide the optional --` + pkgPrefixFlag + ` flag, the graph will not
include packages that have that prefix AND the graph will truncate each
package's name to not include the prefix.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// importTree is a map of "name" -> ["import", "import", ...]
		importTree := tree.NewTree()

		if importsFlags.pkgFilter != "" {
			if err := importTree.SetFilter(importsFlags.pkgFilter); err != nil {
				return err
			}
		}

		if importsFlags.pkgPrefix != "" {
			importTree.SetPrefix(importsFlags.pkgPrefix)
		}

		for _, pkg := range args {
			if err := importTree.Add(pkg, true); err != nil {
				return err
			}
		}

		graph, err := importTree.Graphviz()
		if err != nil {
			return err
		}

		fmt.Println(graph)

		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&importsFlags.pkgFilter, pkgFilterFlag, "", "Regular expression filter to apply to the package list")
	Cmd.Flags().StringVar(&importsFlags.pkgPrefix, pkgPrefixFlag, "", "Prefix filter to apply to the package list")
}

package imports

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spilliams/goraffe/cli/tree"

	"github.com/spf13/cobra"
)

const (
	filterFlag    = "filter"
	prefixFlag    = "prefix"
	singleFlag    = "single"
	highlightFlag = "highlight"

	highlightSingleParent = "single-parent"
)

var importsFlags struct {
	filter    string
	prefix    string
	single    string
	highlight string
}

var Cmd = &cobra.Command{
	Use:   "imports <packages>",
	Short: "Visualize package imports",
	Long: `Visualize package imports.
	
The packages you list as arguments to this command all get added as "roots" in
the graph.

The graph will include everything the roots import, recursively. The graph will
include the entire dependency chain of the packages you list.

If you provide the optional --` + filterFlag + ` flag, the graph will not
include packages that don't match that filter.

If you provide the optional --` + prefixFlag + ` flag, the graph will not
include packages that have that prefix AND the graph will truncate each
package's name to not include the prefix.

If you provide the optional --` + singleFlag + ` flag, the graph will
contain that package, its direct ancestors, and its direct descendants.

If you provide the optional --` + highlightFlag + ` flag, the graph will
highlight matching nodes. The allowed values for this flag are

	` + highlightSingleParent + `		Packages that are only imported by one other package are "` + highlightSingleParent + `"
	
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// importTree is a map of "name" -> ["import", "import", ...]
		importTree := tree.NewTree()

		if importsFlags.filter != "" {
			logrus.Debugf("%s is '%s'", filterFlag, importsFlags.filter)
			if err := importTree.SetFilter(importsFlags.filter); err != nil {
				return err
			}
		}

		if importsFlags.prefix != "" {
			logrus.Debugf("%s is '%s'", prefixFlag, importsFlags.prefix)
			importTree.SetPrefix(importsFlags.prefix)
		}

		for _, pkg := range args {
			if err := importTree.Add(pkg, true); err != nil {
				return err
			}
		}

		if importsFlags.single != "" {
			logrus.Debugf("%s is '%s'", singleFlag, importsFlags.single)
			if err := importTree.Keep(importsFlags.single); err != nil {
				return err
			}
			importTree.Grow(1)
			importTree.Prune()
		}

		if importsFlags.highlight != "" {
			if err := validateHighlight(importsFlags.highlight); err != nil {
				return err
			}
			importTree.SetHighlight(importsFlags.highlight)
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

func validateHighlight(hl string) error {
	if hl == highlightSingleParent {
		return nil
	}
	return fmt.Errorf("unrecognized value for --%s: %s", highlightFlag, hl)
}

func init() {
	Cmd.Flags().StringVar(&importsFlags.filter, filterFlag, "", "Regular expression filter to apply to the package list")
	Cmd.Flags().StringVar(&importsFlags.prefix, prefixFlag, "", "Prefix filter to apply to the package list")
	Cmd.Flags().StringVar(&importsFlags.single, singleFlag, "", "Pick a single package to show ancestors and descendants of")
	Cmd.Flags().StringVar(&importsFlags.highlight, highlightFlag, "", "Highlight certain packages")
}

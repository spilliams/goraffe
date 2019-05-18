package imports

import (
	"github.com/spilliams/goraffe/cli/tree"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var importsFlags struct {
	pkgFilter string
}

var Cmd = &cobra.Command{
	Use: "imports <packages>",
	RunE: func(cmd *cobra.Command, args []string) error {
		// importTree is a map of "name" -> ["import", "import", ...]
		importTree, err := tree.NewTree(importsFlags.pkgFilter)
		if err != nil {
			return err
		}

		for _, pkg := range args {
			if err = importTree.FindImport(pkg); err != nil {
				return err
			}
		}

		logrus.Info(importTree.Graphviz())

		return nil
	},
}

func init() {
	Cmd.Flags().StringVar(&importsFlags.pkgFilter, "filter", ".*", "Regular expression filter to apply to the package list")
}

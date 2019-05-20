package imports

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spilliams/goraffe/cli/tree"

	"github.com/spf13/cobra"
)

// the names of the flags
const (
	filterFlag = "filter"
	prefixFlag = "prefix"
	singleFlag = "single"
	growFlag   = "grow"
	testsFlag  = "includeTests"
)

var flagsDocs = map[string]string{
	filterFlag: "the graph will not include packages that don't match that filter",
	prefixFlag: "the graph will not include packages that have that prefix AND the graph\nwill truncate each package's name to exclude the prefix",
	singleFlag: "the graph will contain that package, its direct ancestors, and its\ndirect descendants",
	growFlag:   "used in conjunction with --" + singleFlag + ", the graph will \"grow\"\nthe tree the number of times you specify",
	testsFlag:  "the tree will include imports from Go test files",
}

var importsFlags struct {
	filter string
	prefix string
	single string
	grow   int
	tests  bool
}

func optionsDoc() string {
	options := []string{filterFlag, prefixFlag, singleFlag, growFlag, testsFlag}
	sort.Strings(options)
	r := ""
	for _, option := range options {
		doc := flagsDocs[option]
		docLines := strings.Split(doc, "\n")
		doc = strings.Join(docLines, "\n\t\t")
		r += fmt.Sprintf("\t--%s\n\t\t%s\n", option, doc)
	}
	return r
}

var Cmd = &cobra.Command{
	Use:   "imports <packages>",
	Short: "Visualize package imports",
	Long: `Visualize package imports.
	
The packages you list as arguments to this command all get added as "roots" in
the graph.

The graph will include everything the roots import, recursively. The graph will
include the entire dependency chain of the packages you list.

Options:
` + optionsDoc() + `

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
			if _, err := importTree.AddRecursive(pkg, importsFlags.tests); err != nil {
				return err
			}
		}

		if importsFlags.single != "" {
			logrus.Debugf("%s is '%s'", singleFlag, importsFlags.single)
			if err := importTree.Keep(importsFlags.single); err != nil {
				return err
			}
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

func init() {
	Cmd.Flags().StringVar(&importsFlags.filter, filterFlag, "", "Regular expression filter to apply to the package list")
	Cmd.Flags().StringVar(&importsFlags.prefix, prefixFlag, "", "Prefix filter to apply to the package list")
	Cmd.Flags().StringVar(&importsFlags.single, singleFlag, "", "Pick a single package to show ancestors and descendants of")
	Cmd.Flags().IntVar(&importsFlags.grow, growFlag, 1, "How far to \"grow\" the tree away from any kept packages")
	Cmd.Flags().BoolVar(&importsFlags.tests, testsFlag, false, "Whether to include imports from Go test files")
}

package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spilliams/goraffe/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Info())
		},
	}
}

package cmd

import (
	"fmt"
	"os"

	"github.com/spilliams/goraffe/internal/version"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var verbose bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "goraffe <command>",
	Short:   "A tool for graphing go packages",
	Version: version.Info(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initLogger)

	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	RootCmd.AddCommand(importsCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(referrersCmd)
}

func initLogger() {
	logrus.SetLevel(logrus.InfoLevel)
	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logrus.SetOutput(RootCmd.OutOrStderr())
	logrus.StandardLogger().Formatter.(*logrus.TextFormatter).DisableTimestamp = true
	logrus.StandardLogger().Formatter.(*logrus.TextFormatter).DisableLevelTruncation = true
}

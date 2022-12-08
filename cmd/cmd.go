package cmd

import (
	"github.com/dreamerjackson/crawler/cmd/master"
	"github.com/dreamerjackson/crawler/cmd/worker"
	"github.com/dreamerjackson/crawler/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print version.",
	Long:  "print version.",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		version.Printer()
	},
}

func Execute() {
	var rootCmd = &cobra.Command{Use: "crawler"}
	rootCmd.AddCommand(master.MasterCmd, worker.WorkerCmd, versionCmd)
	rootCmd.Execute()
}

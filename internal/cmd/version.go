package cmd

import (
	"fmt"

	"gitlab.com/kinnalru/syncerman/internal/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Version prints the version number of Syncerman.`,
	Run: func(cmd *cobra.Command, args []string) {
		v := version.GetVersion()
		if v == "dev" {
			fmt.Printf("Syncerman version %s (%s)\n", v, version.GetGitCommit())
		} else {
			fmt.Printf("Syncerman version %s\n", v)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

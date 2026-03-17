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
		if version.Version == "dev" || version.Version == "" {
			fmt.Printf("Syncerman version %s (commit: %s)\n", version.Version, version.GitCommit)
		} else {
			fmt.Printf("Syncerman version %s\n", version.Version)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

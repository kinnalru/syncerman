package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Version prints the version number of Syncerman.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Syncerman version 0.1.0")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

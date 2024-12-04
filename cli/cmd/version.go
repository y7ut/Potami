package cmd

import (
	"github.com/spf13/cobra"
)

// VersionCmd represents the version command
var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "the version number of potactl",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println(cmd.Parent().Version)
	},
}

func init() {
	RootCmd.AddCommand(VersionCmd)
}

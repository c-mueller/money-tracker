package cmd

import (
	"fmt"

	"icekalt.dev/money-tracker/internal/buildinfo"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(buildinfo.String())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

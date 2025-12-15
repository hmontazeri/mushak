package cli

import (
	"fmt"

	"github.com/hmontazeri/mushak/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mushak version %s\n", version.GetVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

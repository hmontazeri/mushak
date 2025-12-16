package cli

import (
	"fmt"

	"github.com/hmontazeri/mushak/internal/ui"
	"github.com/hmontazeri/mushak/pkg/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		ui.PrintInfo(fmt.Sprintf("mushak version %s", version.GetVersion()))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

package cli

import (
	"fmt"
	"time"

	"github.com/hmontazeri/mushak/internal/ui"
	"github.com/spf13/cobra"
)

// withTimer wraps a cobra command run function and prints execution duration
func withTimer(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		start := time.Now()
		err := fn(cmd, args)
		if err == nil {
			duration := time.Since(start)
			// Format duration to be human readable (e.g. 1.2s)
			// Using Round to avoid excessive precision
			ui.PrintInfo(fmt.Sprintf("%s completed in %v", cmd.Name(), duration.Round(time.Millisecond)))
		}
		return err
	}
}

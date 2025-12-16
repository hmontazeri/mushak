package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mushak",
	Short: "Zero-config, zero-downtime deployments to Linux servers",
	Long: `Mushak is a CLI tool that enables PaaS-like deployments to your own Linux server.
It uses Git push deployments with automatic Docker builds, health checks, and zero-downtime
switching via Caddy reverse proxy.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

var checkUpdateFunc func()

func init() {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Don't check updates for completion, help, or upgrade commands
		if cmd.Name() == "completion" || cmd.Name() == "help" || cmd.Name() == "upgrade" || cmd.Name() == "update" {
			return
		}
		checkUpdateFunc = CheckUpdateAsync()
	}

	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		if checkUpdateFunc != nil {
			checkUpdateFunc()
		}
	}

	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mushak.yaml)")
}

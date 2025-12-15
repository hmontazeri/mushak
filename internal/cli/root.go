package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mushak",
	Short: "Zero-config, zero-downtime deployments to Ubuntu servers",
	Long: `Mushak is a CLI tool that enables PaaS-like deployments to your own Ubuntu server.
It uses Git push deployments with automatic Docker builds, health checks, and zero-downtime
switching via Caddy reverse proxy.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mushak.yaml)")
}

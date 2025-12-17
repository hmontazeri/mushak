package cli

import (
	"fmt"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/server"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/hmontazeri/mushak/internal/ui"
	"github.com/spf13/cobra"
)

var redeployCmd = &cobra.Command{
	Use:   "redeploy",
	Short: "Trigger a redeployment on structure",
	Long: `Trigger a redeployment of the current version on the server.
This is useful if you want to restart the application or apply environment changes
without pushing new code.`,
	RunE: withTimer(runRedeploy),
}

func init() {
	rootCmd.AddCommand(redeployCmd)
}

func runRedeploy(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.LoadDeployConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nHave you run 'mushak init'?", err)
	}

	ui.PrintHeader("Mushak Redeploy")
	ui.PrintKeyValue("Server", fmt.Sprintf("%s@%s", cfg.User, cfg.Host))
	ui.PrintKeyValue("App", cfg.AppName)
	println()

	// Connect SSH
	ui.PrintInfo("Connecting to server...")
	client, err := ssh.NewClient(ssh.Config{
		Host: cfg.Host,
		User: cfg.User,
	})
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer client.Close()
	ui.PrintSuccess("Connected to server")

	executor := ssh.NewExecutor(client)

	// Trigger Redeploy
	if err := server.TriggerRedeploy(executor, cfg); err != nil {
		return err
	}

	return nil
}

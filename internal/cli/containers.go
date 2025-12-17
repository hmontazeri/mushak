package cli

import (
	"fmt"
	"strings"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/hmontazeri/mushak/internal/ui"
	"github.com/spf13/cobra"
)

var containersCmd = &cobra.Command{
	Use:   "containers",
	Short: "List running containers for the application",
	Long: `List all running Docker containers associated with the application.
This is useful to discover container names for use with 'mushak logs --container'.`,
	RunE: withTimer(runContainers),
}

func init() {
	rootCmd.AddCommand(containersCmd)
}

func runContainers(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.LoadDeployConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nHave you run 'mushak init'?", err)
	}

	ui.PrintHeader("Mushak Containers")
	ui.PrintKeyValue("Server", fmt.Sprintf("%s@%s", cfg.User, cfg.Host))
	ui.PrintKeyValue("App", cfg.AppName)
	println()

	// Connect SSH
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

	executor := ssh.NewExecutor(client)

	// List containers matching the app name
	// Format: NAME | STATUS | PORTS
	dockerCmd := fmt.Sprintf(
		"docker ps --filter 'name=%s' --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'",
		cfg.AppName,
	)

	result, err := executor.Run(dockerCmd)
	if err != nil {
		return fmt.Errorf("failed to list containers: %w", err)
	}

	result = strings.TrimSpace(result)
	if result == "" || result == "NAMES\tSTATUS\tPORTS" {
		ui.PrintWarning("No running containers found")
		return nil
	}

	fmt.Println(result)

	return nil
}

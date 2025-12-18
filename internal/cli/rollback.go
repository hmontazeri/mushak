package cli

import (
	"fmt"
	"strings"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/server"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/hmontazeri/mushak/internal/ui"
	"github.com/hmontazeri/mushak/internal/utils"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback [sha]",
	Short: "Rollback to a previous deployment",
	Long: `Rollback to a previous deployment version.

Without arguments, lists available versions for rollback.
With a SHA argument, rolls back to that specific version.

Examples:
  mushak rollback          # List available versions
  mushak rollback abc123d  # Rollback to specific version
  mushak rollback -1       # Rollback to previous version`,
	RunE: withTimer(runRollback),
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

func runRollback(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.LoadDeployConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nHave you run 'mushak init'?", err)
	}

	ui.PrintHeader("Mushak Rollback")
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
	println()

	executor := ssh.NewExecutor(client)

	// List available versions
	ui.PrintInfo("Fetching available versions...")
	versions, err := server.ListVersions(executor, cfg.AppName)
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	if len(versions) == 0 {
		ui.PrintWarning("No previous versions available for rollback.")
		ui.PrintInfo("Deploy at least once to enable rollback functionality.")
		return nil
	}

	// If no argument, show list and prompt for selection
	if len(args) == 0 {
		return showVersionsAndPrompt(executor, cfg, versions)
	}

	// Handle -1 for "previous version"
	targetSHA := args[0]
	if targetSHA == "-1" {
		// Find the first non-current version
		for _, v := range versions {
			if !v.IsCurrent {
				targetSHA = v.SHA
				break
			}
		}
		if targetSHA == "-1" {
			return fmt.Errorf("no previous version available to rollback to")
		}
	}

	// Verify target exists
	var targetVersion *server.DeploymentVersion
	for i, v := range versions {
		if strings.HasPrefix(v.SHA, targetSHA) {
			targetVersion = &versions[i]
			break
		}
	}

	if targetVersion == nil {
		return fmt.Errorf("version %s not found or image not available", targetSHA)
	}

	if targetVersion.IsCurrent {
		ui.PrintWarning(fmt.Sprintf("Version %s is already the current deployment.", targetSHA))
		return nil
	}

	// Execute rollback
	return server.ExecuteRollback(executor, cfg, targetVersion.SHA)
}

func showVersionsAndPrompt(executor *ssh.Executor, cfg *config.DeployConfig, versions []server.DeploymentVersion) error {
	println()
	ui.PrintInfo("Available versions for rollback:")
	println()

	// Print header
	fmt.Printf("  %-10s %-22s %-8s %s\n", "SHA", "DEPLOYED", "STATUS", "")
	fmt.Printf("  %-10s %-22s %-8s %s\n", "---", "--------", "------", "")

	for i, v := range versions {
		status := ""
		marker := ""
		if v.IsCurrent {
			status = "current"
			marker = " ←"
		}

		// Format timestamp nicely
		timestamp := v.Timestamp
		if len(timestamp) > 19 {
			timestamp = timestamp[:19]
		}
		timestamp = strings.Replace(timestamp, "T", " ", 1)

		fmt.Printf("  %-10s %-22s %-8s%s\n", v.SHA, timestamp, status, marker)

		// Only show first 5 versions in list
		if i >= 4 {
			remaining := len(versions) - 5
			if remaining > 0 {
				fmt.Printf("  ... and %d more versions\n", remaining)
			}
			break
		}
	}
	println()

	// Prompt for selection
	targetSHA, err := utils.PromptString("Enter SHA to rollback to (or 'q' to cancel)", "")
	if err != nil {
		return err
	}

	if targetSHA == "q" || targetSHA == "" {
		ui.PrintInfo("Rollback cancelled.")
		return nil
	}

	// Find matching version
	var targetVersion *server.DeploymentVersion
	for i, v := range versions {
		if strings.HasPrefix(v.SHA, targetSHA) {
			targetVersion = &versions[i]
			break
		}
	}

	if targetVersion == nil {
		return fmt.Errorf("version %s not found", targetSHA)
	}

	if targetVersion.IsCurrent {
		ui.PrintWarning(fmt.Sprintf("Version %s is already the current deployment.", targetSHA))
		return nil
	}

	// Confirm rollback
	confirm, err := utils.Confirm(fmt.Sprintf("Rollback to version %s?", targetVersion.SHA))
	if err != nil {
		return err
	}

	if !confirm {
		ui.PrintInfo("Rollback cancelled.")
		return nil
	}

	println()
	return server.ExecuteRollback(executor, cfg, targetVersion.SHA)
}


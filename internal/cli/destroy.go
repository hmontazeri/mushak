package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/server"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/hmontazeri/mushak/internal/ui"
	"github.com/hmontazeri/mushak/internal/utils"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy an app and remove it from the server",
	Long: `Destroy removes an app completely from the server:
- Stops and removes all containers
- Deletes the Git repository
- Removes deployment files
- Removes Caddy configuration
- Removes local Git remote

WARNING: This action is irreversible!`,
	RunE: withTimer(runDestroy),
}

var (
	destroyHost  string
	destroyUser  string
	destroyApp   string
	destroyKey   string
	destroyPort  string
	destroyForce bool
)

func init() {
	rootCmd.AddCommand(destroyCmd)

	destroyCmd.Flags().StringVar(&destroyHost, "host", "", "Server hostname or IP")
	destroyCmd.Flags().StringVar(&destroyUser, "user", "", "SSH username")
	destroyCmd.Flags().StringVar(&destroyApp, "app", "", "App name")
	destroyCmd.Flags().StringVar(&destroyKey, "key", "", "SSH key path")
	destroyCmd.Flags().StringVar(&destroyPort, "port", "22", "SSH port")
	destroyCmd.Flags().BoolVar(&destroyForce, "force", false, "Skip confirmation prompt")
}

func runDestroy(cmd *cobra.Command, args []string) error {
	var cfg *config.DeployConfig
	var err error

	// Try to load config from .mushak/mushak.yaml
	if destroyHost == "" || destroyUser == "" || destroyApp == "" {
		cfg, err = config.LoadDeployConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w\nPlease provide --host, --user, and --app flags", err)
		}

		// Use loaded config
		if destroyHost == "" {
			destroyHost = cfg.Host
		}
		if destroyUser == "" {
			destroyUser = cfg.User
		}
		if destroyApp == "" {
			destroyApp = cfg.AppName
		}
	}

	ui.PrintHeader("Mushak Destroy")
	ui.PrintWarning(fmt.Sprintf("WARNING: This will permanently delete app '%s' from %s@%s", destroyApp, destroyUser, destroyHost))
	println()
	ui.PrintBox([]string{
		"This will:",
		"  - Stop and remove all containers",
		"  - Delete the Git repository",
		"  - Delete all deployment files",
		"  - Remove Caddy configuration",
		"  - Remove local Git remote",
		"",
		"WARNING: THIS ACTION CANNOT BE UNDONE!",
	})

	// Confirm destruction
	if !destroyForce {
		fmt.Printf("\n→ To confirm, type the app name exactly: %s\n", destroyApp)
		confirmed, err := utils.ConfirmDanger(
			"Are you absolutely sure?",
			destroyApp,
		)
		if err != nil {
			return err
		}

		if !confirmed {
			ui.PrintError("App name did not match. Destroy cancelled.")
			return nil
		}
	}

	println()

	// Create SSH client
	ui.PrintInfo("Connecting to server...")
	sshClient, err := ssh.NewClient(ssh.Config{
		Host:    destroyHost,
		Port:    destroyPort,
		User:    destroyUser,
		KeyPath: destroyKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	if err := sshClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer sshClient.Close()

	ui.PrintSuccess("Connected to server")

	executor := ssh.NewExecutor(sshClient)

	// Stop and remove containers
	ui.PrintInfo("Stopping containers...")

	// Stop all mushak containers for this app
	stopCmd := fmt.Sprintf("docker ps -a --format '{{.Names}}' | grep '^mushak-%s-' | xargs -r docker stop", destroyApp)
	if _, err := executor.Run(stopCmd); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to stop containers: %v", err))
	}

	removeCmd := fmt.Sprintf("docker ps -a --format '{{.Names}}' | grep '^mushak-%s-' | xargs -r docker rm", destroyApp)
	if _, err := executor.Run(removeCmd); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to remove containers: %v", err))
	}

	ui.PrintSuccess("Containers stopped and removed")

	// Remove Git repository
	ui.PrintInfo("Removing Git repository...")
	repoPath := fmt.Sprintf("/var/repo/%s.git", destroyApp)
	if _, err := executor.RunSudo(fmt.Sprintf("rm -rf %s", repoPath)); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to remove repo: %v", err))
	}
	ui.PrintSuccess("Git repository removed")

	// Remove deployment files
	ui.PrintInfo("Removing deployment files...")
	deployPath := fmt.Sprintf("/var/www/%s", destroyApp)
	if _, err := executor.RunSudo(fmt.Sprintf("rm -rf %s", deployPath)); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to remove deployment files: %v", err))
	}
	ui.PrintSuccess("Deployment files removed")

	// Remove Caddy configuration
	if err := server.RemoveAppCaddyConfig(executor, destroyApp); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to remove Caddy config: %v", err))
	}
	println()

	// Remove local Git remote
	if cfg != nil && cfg.RemoteName != "" {
		ui.PrintInfo(fmt.Sprintf("Removing Git remote '%s'...", cfg.RemoteName))
		removeRemoteCmd := exec.Command("git", "remote", "remove", cfg.RemoteName)
		if err := removeRemoteCmd.Run(); err != nil {
			ui.PrintWarning(fmt.Sprintf("Failed to remove git remote: %v", err))
		} else {
			ui.PrintSuccess(fmt.Sprintf("Git remote '%s' removed", cfg.RemoteName))
		}
		println()
	}

	// Remove local .mushak directory
	ui.PrintInfo("Removing local configuration...")
	if err := os.RemoveAll(".mushak"); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to remove .mushak directory: %v", err))
	}
	ui.PrintSuccess("Local configuration removed")

	println()
	ui.PrintSeparator()
	ui.PrintSuccess("App Successfully Destroyed")
	ui.PrintSeparator()
	ui.PrintInfo(fmt.Sprintf("App '%s' has been completely removed from the server.", destroyApp))

	return nil
}

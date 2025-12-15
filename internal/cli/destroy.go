package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/server"
	"github.com/hmontazeri/mushak/internal/ssh"
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
	RunE: runDestroy,
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

	fmt.Println("\n=== Mushak Destroy ===")
	fmt.Printf("⚠ WARNING: This will permanently delete app '%s' from %s@%s\n", destroyApp, destroyUser, destroyHost)
	fmt.Println("\nThis will:")
	fmt.Println("  - Stop and remove all containers")
	fmt.Println("  - Delete the Git repository")
	fmt.Println("  - Delete all deployment files")
	fmt.Println("  - Remove Caddy configuration")
	fmt.Println("  - Remove local Git remote")
	fmt.Println("\n⚠ THIS ACTION CANNOT BE UNDONE!")

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
			fmt.Println("\n✗ App name did not match. Destroy cancelled.")
			return nil
		}
	}

	fmt.Println()

	// Create SSH client
	fmt.Println("→ Connecting to server...")
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

	fmt.Println("✓ Connected to server")

	executor := ssh.NewExecutor(sshClient)

	// Stop and remove containers
	fmt.Println("→ Stopping containers...")

	// Stop all mushak containers for this app
	stopCmd := fmt.Sprintf("docker ps -a --format '{{.Names}}' | grep '^mushak-%s-' | xargs -r docker stop", destroyApp)
	if _, err := executor.Run(stopCmd); err != nil {
		fmt.Printf("⚠ Warning: Failed to stop containers: %v\n", err)
	}

	removeCmd := fmt.Sprintf("docker ps -a --format '{{.Names}}' | grep '^mushak-%s-' | xargs -r docker rm", destroyApp)
	if _, err := executor.Run(removeCmd); err != nil {
		fmt.Printf("⚠ Warning: Failed to remove containers: %v\n", err)
	}

	fmt.Println("✓ Containers stopped and removed")

	// Remove Git repository
	fmt.Println("→ Removing Git repository...")
	repoPath := fmt.Sprintf("/var/repo/%s.git", destroyApp)
	if _, err := executor.RunSudo(fmt.Sprintf("rm -rf %s", repoPath)); err != nil {
		fmt.Printf("⚠ Warning: Failed to remove repo: %v\n", err)
	}
	fmt.Println("✓ Git repository removed")

	// Remove deployment files
	fmt.Println("→ Removing deployment files...")
	deployPath := fmt.Sprintf("/var/www/%s", destroyApp)
	if _, err := executor.RunSudo(fmt.Sprintf("rm -rf %s", deployPath)); err != nil {
		fmt.Printf("⚠ Warning: Failed to remove deployment files: %v\n", err)
	}
	fmt.Println("✓ Deployment files removed")

	// Remove Caddy configuration
	if err := server.RemoveAppCaddyConfig(executor, destroyApp); err != nil {
		fmt.Printf("⚠ Warning: Failed to remove Caddy config: %v\n", err)
	}
	fmt.Println()

	// Remove local Git remote
	if cfg != nil && cfg.RemoteName != "" {
		fmt.Printf("→ Removing Git remote '%s'...\n", cfg.RemoteName)
		removeRemoteCmd := exec.Command("git", "remote", "remove", cfg.RemoteName)
		if err := removeRemoteCmd.Run(); err != nil {
			fmt.Printf("⚠ Warning: Failed to remove git remote: %v\n", err)
		} else {
			fmt.Printf("✓ Git remote '%s' removed\n", cfg.RemoteName)
		}
		fmt.Println()
	}

	// Remove local .mushak directory
	fmt.Println("→ Removing local configuration...")
	if err := os.RemoveAll(".mushak"); err != nil {
		fmt.Printf("⚠ Warning: Failed to remove .mushak directory: %v\n", err)
	}
	fmt.Println("✓ Local configuration removed")

	fmt.Println("========================================")
	fmt.Println("✓ App Successfully Destroyed")
	fmt.Println("========================================")
	fmt.Printf("App '%s' has been completely removed from the server.\n", destroyApp)

	return nil
}

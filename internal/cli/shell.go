package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	shellKey string
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Open an interactive shell in the application container",
	Long: `Open an interactive bash/shell session directly to the running application container.
This allows you to inspect the container environment, run commands, and debug issues interactively.`,
	RunE: runShell,
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.Flags().StringVar(&shellKey, "key", "", "SSH key path (default: ~/.ssh/id_rsa)")
}

func runShell(cmd *cobra.Command, args []string) error {
	// Load deployment configuration
	cfg, err := config.LoadDeployConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nHave you run 'mushak init'?", err)
	}

	fmt.Printf("→ Connecting to %s@%s...\n", cfg.User, cfg.Host)

	// Create SSH client
	client, err := ssh.NewClient(ssh.Config{
		Host:    cfg.Host,
		User:    cfg.User,
		KeyPath: shellKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer client.Close()

	executor := ssh.NewExecutor(client)

	// Find the running container for this app
	findCmd := fmt.Sprintf("docker ps --filter 'name=mushak-%s-' --format '{{.ID}}' | head -n 1", cfg.AppName)
	containerID, err := executor.Run(findCmd)
	if err != nil {
		return fmt.Errorf("failed to find running container: %w", err)
	}

	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return fmt.Errorf("no running container found for app '%s'", cfg.AppName)
	}

	fmt.Printf("→ Opening shell in container %s...\n", containerID)

	// Set terminal to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	// Run interactive shell
	dockerCmd := fmt.Sprintf("docker exec -it %s /bin/bash", containerID)
	// Fallback to sh if bash fails? For now let's assume bash is available or let it fail.
	// Users can typically control the base image.

	if err := executor.RunInteractive(dockerCmd, os.Stdin, os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("interactive session ended with error: %w", err)
	}

	return nil
}

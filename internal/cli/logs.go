package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/spf13/cobra"
)

var (
	logsTail   string
	logsFollow bool
	logsKey    string
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Stream logs from the application",
	Long: `Logs streams the standard output and error from the running application container.
By default, it streams logs from the currently deployed version.`,
	RunE: runLogs,
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().StringVarP(&logsTail, "tail", "n", "100", "Number of lines to show from the end of the logs")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", true, "Follow log output")
	logsCmd.Flags().StringVar(&logsKey, "key", "", "SSH key path (default: ~/.ssh/id_rsa)")
}

func runLogs(cmd *cobra.Command, args []string) error {
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
		KeyPath: logsKey,
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
	// The container name format depends on build method (compose vs dockerfile)
	// But it generally starts with "mushak-<AppName>"
	// We want the most recently created one
	findCmd := fmt.Sprintf("docker ps --filter 'name=mushak-%s-' --format '{{.ID}}' | head -n 1", cfg.AppName)
	containerID, err := executor.Run(findCmd)
	if err != nil {
		return fmt.Errorf("failed to find running container: %w", err)
	}

	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return fmt.Errorf("no running container found for app '%s'", cfg.AppName)
	}

	// Build docker logs command
	logArgs := []string{"docker", "logs"}
	if logsTail != "" {
		logArgs = append(logArgs, "--tail", logsTail)
	}
	if logsFollow {
		logArgs = append(logArgs, "-f")
	}
	logArgs = append(logArgs, containerID)

	dockerCmd := strings.Join(logArgs, " ")
	fmt.Printf("→ Streaming logs from container %s...\n", containerID)
	fmt.Println()

	// Stream logs to stdout
	if err := executor.StreamRun(dockerCmd, os.Stdout, os.Stderr); err != nil {
		// If user interrupts (Ctrl+C), StreamRun might return error, but that's expected.
		// However, remote command failure is also an error.
		return fmt.Errorf("log streaming ended: %w", err)
	}

	return nil
}

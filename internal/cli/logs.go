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
	logsTail      string
	logsFollow    bool
	logsKey       string
	logsContainer string
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Stream logs from the application",
	Long: `Logs streams the standard output and error from the running application container.
By default, it streams logs from the currently deployed version.`,
	RunE: withTimer(runLogs),
}

func init() {
	rootCmd.AddCommand(logsCmd)

	logsCmd.Flags().StringVarP(&logsTail, "tail", "n", "100", "Number of lines to show from the end of the logs")
	logsCmd.Flags().BoolVarP(&logsFollow, "follow", "f", true, "Follow log output")
	logsCmd.Flags().StringVar(&logsKey, "key", "", "SSH key path (default: ~/.ssh/id_rsa)")
	logsCmd.Flags().StringVarP(&logsContainer, "container", "c", "", "Filter logs by container name")
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
	// Try multiple patterns to handle different naming conventions:
	// 1. Mushak's naming: mushak-<app>-<sha>-<service> (new override pattern)
	// 2. Mushak's old naming: mushak-<app>-<sha>-<service>-1
	// 3. Infrastructure: <app>_<service> (e.g., bareagent_postgres)
	// Prioritize the web service container

	patterns := []string{
		fmt.Sprintf("mushak-%s-.*-web$", cfg.AppName), // Mushak override naming with web service
		fmt.Sprintf("mushak-%s-.*-web-", cfg.AppName), // Mushak compose naming with web service
		fmt.Sprintf("mushak-%s-", cfg.AppName),        // Any Mushak container for this app
		fmt.Sprintf("%s[_-]web", cfg.AppName),         // Infrastructure or legacy naming with web
		fmt.Sprintf("%s[_-]", cfg.AppName),            // Any container for app
	}

	if logsContainer != "" {
		// If specific container requested, prioritize it
		// We add patterns to match the container name loosely
		customPatterns := []string{
			logsContainer,                                      // Exact name
			fmt.Sprintf("mushak-%s-.*-%s", cfg.AppName, logsContainer), // Mushak service name
			fmt.Sprintf("%s[_-]%s", cfg.AppName, logsContainer),        // Infra service name
		}
		patterns = append(customPatterns, patterns...)
	}

	var containerID string
	for _, pattern := range patterns {
		findCmd := fmt.Sprintf("docker ps --filter 'name=%s' --format '{{.ID}}' | head -n 1", pattern)
		result, err := executor.Run(findCmd)
		if err == nil && strings.TrimSpace(result) != "" {
			containerID = strings.TrimSpace(result)
			break
		}
	}

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

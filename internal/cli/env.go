package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/hooks"
	"github.com/hmontazeri/mushak/internal/server"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/spf13/cobra"
)

var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment variables",
	Long:  `Manage environment variables for your application.`,
}

var envSetCmd = &cobra.Command{
	Use:   "set [KEY=VALUE]...",
	Short: "Set environment variables and redeploy",
	Long: `Set environment variables for the application and trigger a redeployment/restart.
This will update the .env file on the server and restart the application to apply changes.

Example:
  mushak env set DB_HOST=localhost DB_PORT=5432`,
	Args: cobra.MinimumNArgs(1),
	RunE: runEnvSet,
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envSetCmd)
}

func runEnvSet(cmd *cobra.Command, args []string) error {
	// Parse input args into map
	updates := make(map[string]string)
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid argument: %s. Must be KEY=VALUE", arg)
		}
		updates[parts[0]] = parts[1]
	}

	// Load setup
	cfg, err := config.LoadDeployConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nHave you run 'mushak init'?", err)
	}

	fmt.Println("\n=== Mushak Env Set ===")
	fmt.Printf("Server: %s@%s\n", cfg.User, cfg.Host)
	fmt.Printf("App: %s\n", cfg.AppName)
	fmt.Println()

	// Connect SSH
	fmt.Println("→ Connecting to server...")
	client, err := ssh.NewClient(ssh.Config{
		Host: cfg.Host,
		User: cfg.User,
		// Default port/key handled by NewClient
	})
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer client.Close()
	fmt.Println("✓ Connected to server")

	executor := ssh.NewExecutor(client)

	envPath := fmt.Sprintf("/var/www/%s/.env", cfg.AppName)

	// Read existing .env
	var currentContent string
	if out, err := executor.Run("cat " + envPath); err == nil {
		currentContent = out
	}
	// If it fails, assume file missing and start empty

	newContent := updateEnvFile(currentContent, updates)

	// Write back
	fmt.Println("→ Updating .env file...")
	if err := executor.WriteFileSudo(envPath, newContent); err != nil {
		return fmt.Errorf("failed to write .env: %w", err)
	}
	fmt.Println("✓ Updated .env file")

	// Update deployment hook (to ensure it supports .env)
	fmt.Println("→ Updating deployment hook...")
	hookScript := hooks.GeneratePostReceiveHook(cfg.AppName, cfg.Domain, cfg.Branch)
	if err := server.InstallPostReceiveHook(executor, cfg.AppName, hookScript); err != nil {
		return fmt.Errorf("failed to update hook: %w", err)
	}
	fmt.Println("✓ Updated deployment hook")

	// Trigger Redeploy
	fmt.Println("→ Triggering redeploy...")
	
	// Get SHA of current HEAD on server
	shaCmd := fmt.Sprintf("git --git-dir=/var/repo/%s.git rev-parse HEAD", cfg.AppName)
	sha, err := executor.Run(shaCmd)
	if err != nil {
		return fmt.Errorf("failed to get current deployed SHA: %w", err)
	}
	sha = strings.TrimSpace(sha)
	
	// In some cases git output might have newlines or other noise, be careful.
	if len(sha) < 7 {
		return fmt.Errorf("invalid SHA retrieved from server: %s", sha)
	}

	// Trigger hook
	// We need to run it as the user, but referencing the script which is chmod +x
	redeployCmd := fmt.Sprintf(
		"echo \"%s %s refs/heads/%s\" | /var/repo/%s.git/hooks/post-receive",
		sha, sha, cfg.Branch, cfg.AppName,
	)

	fmt.Println("----------------------------------------")
	if err := executor.StreamRun(redeployCmd, os.Stdout, os.Stderr); err != nil {
		return fmt.Errorf("redeploy failed: %w", err)
	}
	fmt.Println("----------------------------------------")

	return nil
}

// updateEnvFile updates environment variables in the content string
func updateEnvFile(content string, updates map[string]string) string {
	lines := strings.Split(content, "\n")
	// If the file ends with a newline, Split returns an empty string at the end.
	// We want to process that only if it's not the only empty string of an empty file.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	seen := make(map[string]bool)
	var newLines []string

	// Process existing lines
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Keep empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			newLines = append(newLines, line)
			continue
		}

		// Check if line is KEY=VALUE
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			if val, ok := updates[key]; ok {
				newLines = append(newLines, fmt.Sprintf("%s=%s", key, val))
				seen[key] = true
			} else {
				// Keep existing
				newLines = append(newLines, line)
			}
		} else {
			// Not a valid env line (maybe malformed), keep it
			newLines = append(newLines, line)
		}
	}

	// Append new keys
	var newKeys []string
	for k := range updates {
		if !seen[k] {
			newKeys = append(newKeys, k)
		}
	}
	sort.Strings(newKeys)

	for _, k := range newKeys {
		newLines = append(newLines, fmt.Sprintf("%s=%s", k, updates[k]))
	}

	// Join with newlines
	result := strings.Join(newLines, "\n")
	// Ensure single trailing newline
	if result != "" {
		result += "\n"
	}
	
	return result
}

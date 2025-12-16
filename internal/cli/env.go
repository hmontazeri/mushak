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
	"github.com/hmontazeri/mushak/internal/ui"
	"github.com/hmontazeri/mushak/internal/utils"
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
	RunE: withTimer(runEnvSet),
}

var envPushCmd = &cobra.Command{
	Use:   "push [file]",
	Short: "Upload local environment file to server",
	Long: `Upload a local .env file to the server.
If no file is specified, automatically detects and uses .env.prod, .env.production, or .env.

Example:
  mushak env push                    # Auto-detect and upload
  mushak env push .env.production    # Upload specific file`,
	RunE: withTimer(runEnvPush),
}

var envPullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Download environment file from server",
	Long: `Download the environment file from the server to local .env.prod.

Example:
  mushak env pull`,
	RunE: withTimer(runEnvPull),
}

var envDiffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare local and server environment files",
	Long: `Show differences between local and server environment files.

Example:
  mushak env diff`,
	RunE: withTimer(runEnvDiff),
}

func init() {
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envSetCmd)
	envCmd.AddCommand(envPushCmd)
	envCmd.AddCommand(envPullCmd)
	envCmd.AddCommand(envDiffCmd)
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

	ui.PrintHeader("Mushak Env Set")
	ui.PrintKeyValue("Server", fmt.Sprintf("%s@%s", cfg.User, cfg.Host))
	ui.PrintKeyValue("App", cfg.AppName)
	println()

	// Connect SSH
	ui.PrintInfo("Connecting to server...")
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
	ui.PrintSuccess("Connected to server")

	executor := ssh.NewExecutor(client)

	// Try .env.prod first, then .env (prefer .env.prod for production deployments)
	envProdPath := fmt.Sprintf("/var/www/%s/.env.prod", cfg.AppName)
	envPath := fmt.Sprintf("/var/www/%s/.env", cfg.AppName)

	var currentContent string
	var targetPath string

	// Check which env file exists, prefer .env.prod
	if out, err := executor.Run("cat " + envProdPath); err == nil {
		currentContent = out
		targetPath = envProdPath
		ui.PrintInfo("Using existing .env.prod")
	} else if out, err := executor.Run("cat " + envPath); err == nil {
		currentContent = out
		targetPath = envPath
		ui.PrintInfo("Using existing .env")
	} else {
		// Neither exists, create .env.prod by default
		targetPath = envProdPath
		ui.PrintInfo("Creating new .env.prod")
	}

	newContent := updateEnvFile(currentContent, updates)

	// Write back
	ui.PrintInfo("Updating environment file...")
	if err := executor.WriteFileSudo(targetPath, newContent); err != nil {
		return fmt.Errorf("failed to write environment file: %w", err)
	}
	ui.PrintSuccess(fmt.Sprintf("Updated %s", targetPath))

	// Update deployment hook (to ensure it supports .env)
	ui.PrintInfo("Updating deployment hook...")
	hookScript := hooks.GeneratePostReceiveHook(cfg.AppName, cfg.Domain, cfg.Branch, false)
	if err := server.InstallPostReceiveHook(executor, cfg.AppName, hookScript); err != nil {
		return fmt.Errorf("failed to update hook: %w", err)
	}
	ui.PrintSuccess("Updated deployment hook")

	// Trigger Redeploy
	ui.PrintInfo("Triggering redeploy...")
	
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

func runEnvPush(cmd *cobra.Command, args []string) error {
	// Determine which file to upload
	var envFile string
	var err error

	if len(args) > 0 {
		// User specified a file
		envFile = args[0]
		if _, err := os.Stat(envFile); err != nil {
			return fmt.Errorf("file not found: %s", envFile)
		}
	} else {
		// Auto-detect
		envFile, err = utils.DetectLocalEnvFile()
		if err != nil {
			return fmt.Errorf("no environment file found. Tried: .env.prod, .env.production, .env")
		}
	}

	// Load config
	cfg, err := config.LoadDeployConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nHave you run 'mushak init'?", err)
	}

	ui.PrintHeader("Mushak Env Push")
	ui.PrintKeyValue("Server", fmt.Sprintf("%s@%s", cfg.User, cfg.Host))
	ui.PrintKeyValue("App", cfg.AppName)
	ui.PrintKeyValue("Local file", envFile)
	println()

	// Read file
	content, err := os.ReadFile(envFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", envFile, err)
	}

	// Show preview
	count, _ := utils.CountEnvVars(envFile)
	ui.PrintInfo(fmt.Sprintf("Uploading %d variable%s...", count, pluralizeEnv(count)))

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

	// Determine target path
	targetFile := ".env.prod"
	if strings.Contains(envFile, ".env.production") {
		targetFile = ".env.prod"
	} else if envFile == ".env" {
		targetFile = ".env"
	}

	targetPath := fmt.Sprintf("/var/www/%s/%s", cfg.AppName, targetFile)
	if err := executor.WriteFileSudo(targetPath, string(content)); err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	ui.PrintSuccess("Environment file uploaded")
	ui.PrintKeyValue("Target", targetPath)
	println()
	ui.PrintInfo("Run 'mushak deploy' to apply changes")

	return nil
}

func runEnvPull(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.LoadDeployConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nHave you run 'mushak init'?", err)
	}

	ui.PrintHeader("Mushak Env Pull")
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

	// Try .env.prod first, then .env
	envProdPath := fmt.Sprintf("/var/www/%s/.env.prod", cfg.AppName)
	envPath := fmt.Sprintf("/var/www/%s/.env", cfg.AppName)

	var content string
	var sourcePath string

	if out, err := executor.Run(fmt.Sprintf("cat %s", envProdPath)); err == nil {
		content = out
		sourcePath = envProdPath
	} else if out, err := executor.Run(fmt.Sprintf("cat %s", envPath)); err == nil {
		content = out
		sourcePath = envPath
	} else {
		return fmt.Errorf("no environment file found on server")
	}

	ui.PrintInfo(fmt.Sprintf("Downloading from %s...", sourcePath))

	// Write to local .env.prod
	localPath := ".env.prod"
	if err := os.WriteFile(localPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", localPath, err)
	}

	count, _ := utils.CountEnvVars(localPath)
	ui.PrintSuccess(fmt.Sprintf("Downloaded %d variable%s to %s", count, pluralizeEnv(count), localPath))

	return nil
}

func runEnvDiff(cmd *cobra.Command, args []string) error {
	// Detect local env file
	localFile, err := utils.DetectLocalEnvFile()
	if err != nil {
		return fmt.Errorf("no local environment file found")
	}

	// Load config
	cfg, err := config.LoadDeployConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nHave you run 'mushak init'?", err)
	}

	ui.PrintHeader("Mushak Env Diff")
	ui.PrintKeyValue("Local", localFile)
	ui.PrintKeyValue("Server", fmt.Sprintf("%s@%s (%s)", cfg.User, cfg.Host, cfg.AppName))
	println()

	// Read local
	localVars, err := utils.ParseEnvFile(localFile)
	if err != nil {
		return fmt.Errorf("failed to parse local file: %w", err)
	}

	// Connect and read remote
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

	// Try .env.prod first, then .env
	envProdPath := fmt.Sprintf("/var/www/%s/.env.prod", cfg.AppName)
	envPath := fmt.Sprintf("/var/www/%s/.env", cfg.AppName)

	var remoteContent string
	if out, err := executor.Run(fmt.Sprintf("cat %s", envProdPath)); err == nil {
		remoteContent = out
	} else if out, err := executor.Run(fmt.Sprintf("cat %s", envPath)); err == nil {
		remoteContent = out
	} else {
		return fmt.Errorf("no environment file found on server")
	}

	// Parse remote
	remoteVars := make(map[string]string)
	for _, line := range strings.Split(remoteContent, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			remoteVars[parts[0]] = parts[1]
		}
	}

	// Compare
	allKeys := make(map[string]bool)
	for k := range localVars {
		allKeys[k] = true
	}
	for k := range remoteVars {
		allKeys[k] = true
	}

	keys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	hasChanges := false
	for _, key := range keys {
		localVal, localExists := localVars[key]
		remoteVal, remoteExists := remoteVars[key]

		if !remoteExists {
			fmt.Println(ui.Success(fmt.Sprintf("+ %s (only in local)", key)))
			hasChanges = true
		} else if !localExists {
			fmt.Println(ui.Error(fmt.Sprintf("- %s (only on server)", key)))
			hasChanges = true
		} else if localVal != remoteVal {
			fmt.Println(ui.Warning(fmt.Sprintf("≠ %s (values differ)", key)))
			hasChanges = true
		}
	}

	if !hasChanges {
		ui.PrintSuccess("No differences found")
	} else {
		println()
		ui.PrintInfo("Use 'mushak env push' to upload local changes")
		ui.PrintInfo("Use 'mushak env pull' to download server version")
	}

	return nil
}

func pluralizeEnv(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

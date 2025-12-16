package server

import (
	"fmt"

	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/hmontazeri/mushak/internal/ui"
)

// SetupGitRepo creates a bare Git repository on the server
func SetupGitRepo(executor *ssh.Executor, appName string) error {
	ui.PrintInfo(fmt.Sprintf("Setting up Git repository for %s...", appName))

	repoPath := fmt.Sprintf("/var/repo/%s.git", appName)
	deployPath := fmt.Sprintf("/var/www/%s", appName)

	// Create repo directory
	if _, err := executor.RunSudo(fmt.Sprintf("mkdir -p %s", repoPath)); err != nil {
		return fmt.Errorf("failed to create repo directory: %w", err)
	}

	// Create deploy directory
	if _, err := executor.RunSudo(fmt.Sprintf("mkdir -p %s", deployPath)); err != nil {
		return fmt.Errorf("failed to create deploy directory: %w", err)
	}

	// Initialize bare repository
	if _, err := executor.RunSudo(fmt.Sprintf("git init --bare %s", repoPath)); err != nil {
		return fmt.Errorf("failed to initialize bare repo: %w", err)
	}

	// Set permissions (allow current user to push)
	user, err := executor.Run("whoami")
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	user = user[:len(user)-1] // Remove trailing newline

	if _, err := executor.RunSudo(fmt.Sprintf("chown -R %s:%s %s", user, user, repoPath)); err != nil {
		return fmt.Errorf("failed to set repo permissions: %w", err)
	}

	if _, err := executor.RunSudo(fmt.Sprintf("chown -R %s:%s %s", user, user, deployPath)); err != nil {
		return fmt.Errorf("failed to set deploy permissions: %w", err)
	}

	ui.PrintSuccess("Git repository created")
	return nil
}

// InstallPostReceiveHook installs the post-receive hook
func InstallPostReceiveHook(executor *ssh.Executor, appName, hookScript string) error {
	ui.PrintInfo("Installing post-receive hook...")

	hookPath := fmt.Sprintf("/var/repo/%s.git/hooks/post-receive", appName)

	// Write hook script
	if err := executor.WriteFile(hookPath, hookScript); err != nil {
		return fmt.Errorf("failed to write hook: %w", err)
	}

	// Make hook executable
	if _, err := executor.Run(fmt.Sprintf("chmod +x %s", hookPath)); err != nil {
		return fmt.Errorf("failed to make hook executable: %w", err)
	}

	ui.PrintSuccess("Post-receive hook installed")
	return nil
}

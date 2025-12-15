package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/hooks"
	"github.com/hmontazeri/mushak/internal/server"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new app on the server",
	Long: `Initialize sets up everything needed to deploy an app:
- Installs Docker, Git, and Caddy on the server
- Creates a bare Git repository
- Configures Caddy for multi-app support
- Installs the post-receive deployment hook
- Adds a Git remote to your local repository`,
	RunE: runInit,
}

var (
	initHost   string
	initUser   string
	initDomain string
	initApp    string
	initBranch string
	initKey    string
	initPort   string
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().StringVar(&initHost, "host", "", "Server hostname or IP (required)")
	initCmd.Flags().StringVar(&initUser, "user", "", "SSH username (required)")
	initCmd.Flags().StringVar(&initDomain, "domain", "", "Domain name for the app (required)")
	initCmd.Flags().StringVar(&initApp, "app", "", "App name (default: current directory name)")
	initCmd.Flags().StringVar(&initBranch, "branch", "main", "Git branch to deploy")
	initCmd.Flags().StringVar(&initKey, "key", "", "SSH key path (default: ~/.ssh/id_rsa)")
	initCmd.Flags().StringVar(&initPort, "port", "22", "SSH port")

	initCmd.MarkFlagRequired("host")
	initCmd.MarkFlagRequired("user")
	initCmd.MarkFlagRequired("domain")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Determine app name
	if initApp == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		initApp = filepath.Base(cwd)
		fmt.Printf("Using app name: %s\n", initApp)
	}

	// Validate we're in a git repository
	if !isGitRepo() {
		return fmt.Errorf("not a git repository. Please run 'git init' first")
	}

	fmt.Println("\n=== Mushak Initialization ===")
	fmt.Printf("Server: %s@%s\n", initUser, initHost)
	fmt.Printf("App: %s\n", initApp)
	fmt.Printf("Domain: %s\n", initDomain)
	fmt.Printf("Branch: %s\n", initBranch)
	fmt.Println()

	// Create SSH client
	fmt.Println("→ Connecting to server...")
	sshClient, err := ssh.NewClient(ssh.Config{
		Host:    initHost,
		Port:    initPort,
		User:    initUser,
		KeyPath: initKey,
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

	// Install dependencies
	if err := server.InstallDependencies(executor); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}
	fmt.Println()

	// Initialize Caddy multi-app setup
	if err := server.InitializeCaddyMultiApp(executor); err != nil {
		return fmt.Errorf("failed to initialize Caddy: %w", err)
	}
	fmt.Println()

	// Setup Git repository
	if err := server.SetupGitRepo(executor, initApp); err != nil {
		return fmt.Errorf("failed to setup Git repo: %w", err)
	}
	fmt.Println()

	// Generate and install post-receive hook
	hookScript := hooks.GeneratePostReceiveHook(initApp, initDomain, initBranch)
	if err := server.InstallPostReceiveHook(executor, initApp, hookScript); err != nil {
		return fmt.Errorf("failed to install post-receive hook: %w", err)
	}
	fmt.Println()

	// Create initial Caddy config (placeholder)
	fmt.Println("→ Creating initial Caddy configuration...")
	configPath := fmt.Sprintf("/etc/caddy/apps/%s.caddy", initApp)
	placeholderConfig := fmt.Sprintf(`# %s - Not yet deployed
# This file will be updated after first deployment
`, initApp)

	if err := executor.WriteFileSudo(configPath, placeholderConfig); err != nil {
		return fmt.Errorf("failed to create Caddy config: %w", err)
	}
	fmt.Println("✓ Caddy configuration created")

	// Add Git remote
	remoteName := "mushak"
	remoteURL := fmt.Sprintf("ssh://%s@%s:%s/var/repo/%s.git", initUser, initHost, initPort, initApp)

	fmt.Printf("→ Adding Git remote '%s'...\n", remoteName)

	// Check if remote already exists
	checkCmd := exec.Command("git", "remote", "get-url", remoteName)
	if err := checkCmd.Run(); err == nil {
		// Remote exists, update it
		updateCmd := exec.Command("git", "remote", "set-url", remoteName, remoteURL)
		if err := updateCmd.Run(); err != nil {
			return fmt.Errorf("failed to update Git remote: %w", err)
		}
		fmt.Printf("✓ Updated Git remote '%s'\n", remoteName)
	} else {
		// Add new remote
		addCmd := exec.Command("git", "remote", "add", remoteName, remoteURL)
		if err := addCmd.Run(); err != nil {
			return fmt.Errorf("failed to add Git remote: %w", err)
		}
		fmt.Printf("✓ Added Git remote '%s'\n", remoteName)
	}
	fmt.Println()

	// Save configuration locally
	deployConfig := &config.DeployConfig{
		AppName:    initApp,
		Host:       initHost,
		User:       initUser,
		Domain:     initDomain,
		Branch:     initBranch,
		RemoteName: remoteName,
	}

	if err := config.SaveDeployConfig(deployConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("========================================")
	fmt.Println("✓ Initialization Complete!")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Printf("  1. Deploy your app: mushak deploy\n")
	fmt.Printf("  2. Your app will be available at: https://%s\n", initDomain)
	fmt.Println()
	fmt.Println("Note: Make sure DNS for your domain points to the server")
	fmt.Println("========================================")

	return nil
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/hooks"
	"github.com/hmontazeri/mushak/internal/server"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/hmontazeri/mushak/internal/ui"
	"github.com/hmontazeri/mushak/internal/utils"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [USER@HOST]",
	Short: "Initialize a new app on the server",
	Long: `Initialize sets up everything needed to deploy an app:
- Installs Docker, Git, and Caddy on the server
- Creates a bare Git repository
- Configures Caddy for multi-app support
- Installs the post-receive deployment hook
- Adds a Git remote to your local repository

Usage:
  mushak init USER@HOST

Example:
  mushak init root@192.168.1.100`,
	RunE: withTimer(runInit),
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

	initCmd.Flags().StringVar(&initHost, "host", "", "Server hostname or IP")
	initCmd.Flags().StringVar(&initUser, "user", "", "SSH username")
	initCmd.Flags().StringVar(&initDomain, "domain", "", "Domain name for the app")
	initCmd.Flags().StringVar(&initApp, "app", "", "App name (default: current directory name)")
	initCmd.Flags().StringVar(&initBranch, "branch", "main", "Git branch to deploy")
	initCmd.Flags().StringVar(&initKey, "key", "", "SSH key path (default: ~/.ssh/id_rsa)")
	initCmd.Flags().StringVar(&initPort, "port", "22", "SSH port")
}

func runInit(cmd *cobra.Command, args []string) error {
	// Parse USER@HOST from positional argument if provided
	if len(args) > 0 {
		userHost := args[0]
		parts := splitUserHost(userHost)
		if parts == nil {
			return fmt.Errorf("invalid format. Expected USER@HOST (e.g., root@192.168.1.100)")
		}
		// Only set if not already provided via flags
		if initUser == "" {
			initUser = parts[0]
		}
		if initHost == "" {
			initHost = parts[1]
		}
	}

	// Validate we're in a git repository
	if !isGitRepo() {
		return fmt.Errorf("not a git repository. Please run 'git init' first")
	}

	// Get current directory for default app name
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	defaultApp := filepath.Base(cwd)

	// Prompt for domain if not provided
	if initDomain == "" {
		domain, err := utils.PromptString("Domain", "")
		if err != nil {
			return fmt.Errorf("failed to get domain: %w", err)
		}
		if domain == "" {
			return fmt.Errorf("domain is required")
		}
		initDomain = domain
	}

	// Prompt for app name if not provided (with default)
	if initApp == "" {
		app, err := utils.PromptString("App name", defaultApp)
		if err != nil {
			return fmt.Errorf("failed to get app name: %w", err)
		}
		if app == "" {
			initApp = defaultApp
		} else {
			initApp = app
		}
	}

	// Validate required fields
	if initUser == "" || initHost == "" {
		return fmt.Errorf("user and host are required. Usage: mushak init USER@HOST")
	}

	// Print banner
	ui.PrintBanner()

	// Print configuration
	ui.PrintHeader("Initialization")
	ui.PrintKeyValue("Server", fmt.Sprintf("%s@%s", initUser, initHost))
	ui.PrintKeyValue("App", initApp)
	ui.PrintKeyValue("Domain", initDomain)
	ui.PrintKeyValue("Branch", initBranch)
	println()

	// Create SSH client
	ui.PrintInfo("Connecting to server...")
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

	ui.PrintSuccess("Connected to server")

	executor := ssh.NewExecutor(sshClient)

	// Install dependencies
	if err := server.InstallDependencies(executor); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}
	println()

	// Initialize Caddy multi-app setup
	if err := server.InitializeCaddyMultiApp(executor); err != nil {
		return fmt.Errorf("failed to initialize Caddy: %w", err)
	}
	println()

	// Setup Git repository
	if err := server.SetupGitRepo(executor, initApp); err != nil {
		return fmt.Errorf("failed to setup Git repo: %w", err)
	}
	println()

	// Generate and install post-receive hook
	hookScript := hooks.GeneratePostReceiveHook(initApp, initDomain, initBranch, false)
	if err := server.InstallPostReceiveHook(executor, initApp, hookScript); err != nil {
		return fmt.Errorf("failed to install post-receive hook: %w", err)
	}
	println()

	// Create initial Caddy config (placeholder)
	ui.PrintInfo("Creating initial Caddy configuration...")
	configPath := fmt.Sprintf("/etc/caddy/apps/%s.caddy", initApp)
	placeholderConfig := fmt.Sprintf(`# %s - Not yet deployed
# This file will be updated after first deployment
`, initApp)

	if err := executor.WriteFileSudo(configPath, placeholderConfig); err != nil {
		return fmt.Errorf("failed to create Caddy config: %w", err)
	}
	ui.PrintSuccess("Caddy configuration created")
	println()

	// Check for local env file and prompt to upload
	if envFile, err := detectAndUploadEnvFile(executor, initApp); err == nil && envFile != "" {
		ui.PrintSuccess(fmt.Sprintf("Uploaded %s to server", envFile))
	}

	// Add Git remote
	remoteName := "mushak"
	remoteURL := fmt.Sprintf("ssh://%s@%s:%s/var/repo/%s.git", initUser, initHost, initPort, initApp)

	ui.PrintInfo(fmt.Sprintf("Adding Git remote '%s'...", remoteName))

	// Check if remote already exists
	checkCmd := exec.Command("git", "remote", "get-url", remoteName)
	if err := checkCmd.Run(); err == nil {
		// Remote exists, update it
		updateCmd := exec.Command("git", "remote", "set-url", remoteName, remoteURL)
		if err := updateCmd.Run(); err != nil {
			return fmt.Errorf("failed to update Git remote: %w", err)
		}
		ui.PrintSuccess(fmt.Sprintf("Updated Git remote '%s'", remoteName))
	} else {
		// Add new remote
		addCmd := exec.Command("git", "remote", "add", remoteName, remoteURL)
		if err := addCmd.Run(); err != nil {
			return fmt.Errorf("failed to add Git remote: %w", err)
		}
		ui.PrintSuccess(fmt.Sprintf("Added Git remote '%s'", remoteName))
	}
	println()

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

	// Success message
	println()
	ui.PrintSeparator()
	ui.PrintSuccess("Initialization Complete!")
	ui.PrintSeparator()
	println()
	ui.PrintHeader("Next Steps")
	ui.PrintKeyValue("1", "Deploy your app with 'mushak deploy'")
	ui.PrintKeyValue("2", fmt.Sprintf("Access at https://%s", initDomain))
	println()
	ui.PrintInfo("Make sure DNS for your domain points to the server")
	println()

	return nil
}

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run() == nil
}

// detectAndUploadEnvFile detects local env file and prompts user to upload
func detectAndUploadEnvFile(executor *ssh.Executor, appName string) (string, error) {
	envFile, err := detectLocalEnvFileWithFallback()
	if err != nil {
		// No local env file found, skip silently
		return "", nil
	}

	// Count variables
	count, err := countEnvVars(envFile)
	if err != nil || count == 0 {
		return "", nil
	}

	// Get first few variable names for display
	keys, _ := getEnvVarKeys(envFile)
	preview := ""
	if len(keys) > 0 {
		displayKeys := keys
		if len(keys) > 3 {
			displayKeys = keys[:3]
		}
		preview = fmt.Sprintf(" (%s", displayKeys[0])
		for i := 1; i < len(displayKeys); i++ {
			preview += ", " + displayKeys[i]
		}
		if len(keys) > 3 {
			preview += fmt.Sprintf(", +%d more", len(keys)-3)
		}
		preview += ")"
	}

	ui.PrintSuccess(fmt.Sprintf("Found local %s with %d variable%s%s", envFile, count, pluralize(count), preview))

	// Prompt user
	confirmed, err := confirmEnvUpload()
	if err != nil || !confirmed {
		ui.PrintInfo("Skipped environment file upload")
		return "", nil
	}

	// Read and upload
	content, err := os.ReadFile(envFile)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", envFile, err)
	}

	// Determine target path (use .env.prod if source is .env.prod, otherwise .env)
	targetFile := ".env.prod"
	if envFile == ".env" {
		targetFile = ".env"
	}

	targetPath := fmt.Sprintf("/var/www/%s/%s", appName, targetFile)
	if err := executor.WriteFileSudo(targetPath, string(content)); err != nil {
		return "", fmt.Errorf("failed to upload environment file: %w", err)
	}

	return envFile, nil
}

func detectLocalEnvFileWithFallback() (string, error) {
	return utils.DetectLocalEnvFile()
}

func countEnvVars(path string) (int, error) {
	return utils.CountEnvVars(path)
}

func getEnvVarKeys(path string) ([]string, error) {
	return utils.GetEnvVarKeys(path)
}

func confirmEnvUpload() (bool, error) {
	return utils.Confirm("→ Upload to server?")
}

// splitUserHost parses USER@HOST format and returns [user, host]
func splitUserHost(userHost string) []string {
	parts := strings.Split(userHost, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil
	}
	return parts
}

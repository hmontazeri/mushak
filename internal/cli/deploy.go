package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/hmontazeri/mushak/internal/utils"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy the current branch to the server",
	Long: `Deploy pushes your code to the server's Git repository, which triggers
the post-receive hook to build and deploy your application with zero downtime.`,
	RunE: runDeploy,
}

var deployForce bool

func init() {
	rootCmd.AddCommand(deployCmd)

	deployCmd.Flags().BoolVarP(&deployForce, "force", "f", false, "Force push to server")
}

func runDeploy(cmd *cobra.Command, args []string) error {
	// Load deployment configuration
	cfg, err := config.LoadDeployConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w\nHave you run 'mushak init'?", err)
	}

	// Validate we're in a git repository
	if !isGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	// Get current branch
	currentBranch, err := getCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	fmt.Println("\n=== Mushak Deployment ===")
	fmt.Printf("App: %s\n", cfg.AppName)
	fmt.Printf("Server: %s@%s\n", cfg.User, cfg.Host)
	fmt.Printf("Branch: %s -> %s\n", currentBranch, cfg.Branch)
	fmt.Printf("Domain: https://%s\n", cfg.Domain)
	fmt.Println()

	// Check if current branch matches configured branch
	if currentBranch != cfg.Branch {
		fmt.Printf("⚠ Warning: You're on branch '%s' but configured to deploy '%s'\n", currentBranch, cfg.Branch)
		fmt.Println()
	}

	// Verify git remote exists
	checkCmd := exec.Command("git", "remote", "get-url", cfg.RemoteName)
	if err := checkCmd.Run(); err != nil {
		return fmt.Errorf("git remote '%s' not found. Please run 'mushak init' first", cfg.RemoteName)
	}

	// Check for environment files and prompt if needed
	if err := checkAndUploadEnvFile(cfg); err != nil {
		fmt.Printf("⚠ Warning: %v\n", err)
	}
	fmt.Println()

	// Build push command
	pushArgs := []string{"push", cfg.RemoteName, fmt.Sprintf("HEAD:refs/heads/%s", cfg.Branch)}
	if deployForce {
		pushArgs = append(pushArgs, "--force")
	}

	fmt.Println("→ Pushing to server...")
	fmt.Println()

	// Execute git push with output streaming
	pushCmd := exec.Command("git", pushArgs...)
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	pushCmd.Stdin = os.Stdin

	if err := pushCmd.Run(); err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	return nil
}

func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	branch := string(output)
	// Remove trailing newline
	if len(branch) > 0 && branch[len(branch)-1] == '\n' {
		branch = branch[:len(branch)-1]
	}

	return branch, nil
}

// checkAndUploadEnvFile checks if env file exists on server, if not prompts to upload local
func checkAndUploadEnvFile(cfg *config.DeployConfig) error {
	// Check if local env file exists
	localEnvFile, err := utils.DetectLocalEnvFile()
	if err != nil {
		// No local env file, nothing to do
		return nil
	}

	// Connect to server to check if env file exists
	sshClient, err := ssh.NewClient(ssh.Config{
		Host: cfg.Host,
		User: cfg.User,
	})
	if err != nil {
		return fmt.Errorf("failed to create SSH client: %w", err)
	}

	if err := sshClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer sshClient.Close()

	executor := ssh.NewExecutor(sshClient)

	// Check if .env.prod or .env exists on server
	envProdPath := fmt.Sprintf("/var/www/%s/.env.prod", cfg.AppName)
	envPath := fmt.Sprintf("/var/www/%s/.env", cfg.AppName)

	_, errProd := executor.Run(fmt.Sprintf("test -f %s", envProdPath))
	_, errEnv := executor.Run(fmt.Sprintf("test -f %s", envPath))

	// If both checks fail, env file doesn't exist on server
	if errProd != nil && errEnv != nil {
		// Server has no env file, but we have one locally
		count, _ := utils.CountEnvVars(localEnvFile)
		if count == 0 {
			return nil
		}

		// Get preview of variables
		keys, _ := utils.GetEnvVarKeys(localEnvFile)
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

		fmt.Printf("⚠ No environment file found on server\n")
		fmt.Printf("✓ Found local %s with %d variable%s%s\n", localEnvFile, count, pluralize(count), preview)

		// Prompt user
		confirmed, err := utils.Confirm("→ Upload to server?")
		if err != nil || !confirmed {
			fmt.Println("  Skipped. Use 'mushak env push' to upload later")
			return nil
		}

		// Read and upload
		content, err := os.ReadFile(localEnvFile)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", localEnvFile, err)
		}

		// Determine target (prefer .env.prod)
		targetFile := ".env.prod"
		if localEnvFile == ".env" {
			targetFile = ".env"
		}

		targetPath := fmt.Sprintf("/var/www/%s/%s", cfg.AppName, targetFile)
		if err := executor.WriteFileSudo(targetPath, string(content)); err != nil {
			return fmt.Errorf("failed to upload: %w", err)
		}

		fmt.Printf("✓ Uploaded %s to server\n", localEnvFile)
	} else {
		// Env file exists on server
		if errProd == nil {
			fmt.Println("→ Environment file: .env.prod ✓")
		} else {
			fmt.Println("→ Environment file: .env ✓")
		}
	}

	return nil
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

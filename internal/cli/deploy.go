package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/hmontazeri/mushak/internal/config"
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

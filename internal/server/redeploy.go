package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/hmontazeri/mushak/internal/config"
	"github.com/hmontazeri/mushak/internal/ssh"
)

// TriggerRedeploy triggers a redeployment using the existing code on the server
func TriggerRedeploy(executor *ssh.Executor, cfg *config.DeployConfig) error {
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

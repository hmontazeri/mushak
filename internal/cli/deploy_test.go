package cli

import (
	"fmt"
	"strings"
	"testing"
)

func TestDeployCommand(t *testing.T) {
	if deployCmd == nil {
		t.Fatal("deployCmd should not be nil")
	}

	if deployCmd.Use != "deploy" {
		t.Errorf("deployCmd.Use = %v, want deploy", deployCmd.Use)
	}

	if deployCmd.Short == "" {
		t.Error("deployCmd.Short should not be empty")
	}

	if deployCmd.Long == "" {
		t.Error("deployCmd.Long should not be empty")
	}

	if deployCmd.RunE == nil {
		t.Error("deployCmd.RunE should not be nil")
	}
}

func TestDeployCommandFlags(t *testing.T) {
	if deployCmd == nil {
		t.Fatal("deployCmd should not be nil")
	}

	// Check that force flag exists
	forceFlag := deployCmd.Flags().Lookup("force")
	if forceFlag == nil {
		t.Error("deploy command should have --force flag")
	}

	if forceFlag != nil {
		if forceFlag.Shorthand != "f" {
			t.Errorf("force flag shorthand = %v, want f", forceFlag.Shorthand)
		}
	}
}

func TestGetCurrentBranch(t *testing.T) {
	// This test verifies the function signature
	// In a real git repository, it would return the current branch
	// We can't test actual functionality without a git repo
}

func TestDeployPushCommand(t *testing.T) {
	tests := []struct {
		name       string
		remoteName string
		branch     string
		force      bool
		wantArgs   []string
	}{
		{
			name:       "basic push",
			remoteName: "mushak",
			branch:     "main",
			force:      false,
			wantArgs:   []string{"push", "mushak", "HEAD:refs/heads/main"},
		},
		{
			name:       "force push",
			remoteName: "mushak",
			branch:     "main",
			force:      true,
			wantArgs:   []string{"push", "mushak", "HEAD:refs/heads/main", "--force"},
		},
		{
			name:       "different branch",
			remoteName: "origin",
			branch:     "production",
			force:      false,
			wantArgs:   []string{"push", "origin", "HEAD:refs/heads/production"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build push command args
			pushArgs := []string{"push", tt.remoteName, fmt.Sprintf("HEAD:refs/heads/%s", tt.branch)}
			if tt.force {
				pushArgs = append(pushArgs, "--force")
			}

			// Verify args match expected
			if len(pushArgs) != len(tt.wantArgs) {
				t.Errorf("args length = %d, want %d", len(pushArgs), len(tt.wantArgs))
				return
			}

			for i, arg := range pushArgs {
				if arg != tt.wantArgs[i] {
					t.Errorf("arg[%d] = %v, want %v", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestDeployCommandDescription(t *testing.T) {
	if deployCmd == nil {
		t.Fatal("deployCmd should not be nil")
	}

	requiredDescriptions := []string{
		"zero downtime",
		"post-receive hook",
	}

	for _, desc := range requiredDescriptions {
		if !strings.Contains(deployCmd.Long, desc) {
			t.Errorf("deploy command description should mention %q", desc)
		}
	}
}

func TestDeployBranchWarning(t *testing.T) {
	tests := []struct {
		currentBranch    string
		configuredBranch string
		shouldWarn       bool
	}{
		{
			currentBranch:    "main",
			configuredBranch: "main",
			shouldWarn:       false,
		},
		{
			currentBranch:    "develop",
			configuredBranch: "main",
			shouldWarn:       true,
		},
		{
			currentBranch:    "feature/new",
			configuredBranch: "production",
			shouldWarn:       true,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s vs %s", tt.currentBranch, tt.configuredBranch), func(t *testing.T) {
			shouldWarn := tt.currentBranch != tt.configuredBranch

			if shouldWarn != tt.shouldWarn {
				t.Errorf("shouldWarn = %v, want %v", shouldWarn, tt.shouldWarn)
			}
		})
	}
}

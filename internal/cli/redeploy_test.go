package cli

import (
	"strings"
	"testing"
)

func TestRedeployCommand(t *testing.T) {
	if redeployCmd == nil {
		t.Fatal("redeployCmd should not be nil")
	}

	if redeployCmd.Use != "redeploy" {
		t.Errorf("redeployCmd.Use = %v, want redeploy", redeployCmd.Use)
	}

	if redeployCmd.Short == "" {
		t.Error("redeployCmd.Short should not be empty")
	}

	if redeployCmd.Long == "" {
		t.Error("redeployCmd.Long should not be empty")
	}

	if redeployCmd.RunE == nil {
		t.Error("redeployCmd.RunE should not be nil")
	}
	
	requiredDescriptions := []string{
		"Trigger",
		"redeployment",
	}

	for _, desc := range requiredDescriptions {
		if !strings.Contains(redeployCmd.Long, desc) && !strings.Contains(redeployCmd.Short, desc) {
			t.Errorf("redeploy command description should mention %q", desc)
		}
	}
}

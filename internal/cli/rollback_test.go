package cli

import (
	"testing"
)

func TestRollbackCommand(t *testing.T) {
	// Test command exists and has correct use
	if rollbackCmd.Use != "rollback [sha]" {
		t.Errorf("Expected Use 'rollback [sha]', got '%s'", rollbackCmd.Use)
	}

	// Test short description
	if rollbackCmd.Short != "Rollback to a previous deployment" {
		t.Errorf("Unexpected Short description: %s", rollbackCmd.Short)
	}
}

func TestRollbackCommandExamples(t *testing.T) {
	// Verify the long description contains expected examples
	long := rollbackCmd.Long
	
	expectedExamples := []string{
		"mushak rollback",
		"mushak rollback abc123d",
		"mushak rollback -1",
	}
	
	for _, example := range expectedExamples {
		if !contains(long, example) {
			t.Errorf("Long description should contain example: %s", example)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}


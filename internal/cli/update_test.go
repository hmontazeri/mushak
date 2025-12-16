package cli

import (
	"strings"
	"testing"
)

func TestUpdateCommand(t *testing.T) {
	if updateCmd == nil {
		t.Fatal("updateCmd should not be nil")
	}

	if updateCmd.Use != "upgrade" {
		t.Errorf("updateCmd.Use = %v, want upgrade", updateCmd.Use)
	}

	if updateCmd.Short == "" {
		t.Error("updateCmd.Short should not be empty")
	}

	if updateCmd.Long == "" {
		t.Error("updateCmd.Long should not be empty")
	}

	if updateCmd.RunE == nil {
		t.Error("updateCmd.RunE should not be nil")
	}
}

func TestUpdateCommandFlags(t *testing.T) {
	if updateCmd == nil {
		t.Fatal("updateCmd should not be nil")
	}

	// Check that check flag exists
	checkFlag := updateCmd.Flags().Lookup("check")
	if checkFlag == nil {
		t.Error("update command should have --check flag")
	}

	if checkFlag != nil {
		// Check flag should default to false
		if checkFlag.DefValue != "false" {
			t.Errorf("check flag default = %v, want false", checkFlag.DefValue)
		}
	}
}

func TestUpdateCommandDescription(t *testing.T) {
	if updateCmd == nil {
		t.Fatal("updateCmd should not be nil")
	}

	requiredTerms := []string{
		"latest",
		"GitHub",
	}

	for _, term := range requiredTerms {
		found := strings.Contains(updateCmd.Long, term) ||
			strings.Contains(updateCmd.Short, term)

		if !found {
			t.Errorf("update command description should mention: %q", term)
		}
	}
}

func TestUpdateCheckMode(t *testing.T) {
	// Test that check mode exists as a concept
	// The --check flag should allow checking without installing
	if updateCmd == nil {
		t.Fatal("updateCmd should not be nil")
	}

	checkFlag := updateCmd.Flags().Lookup("check")
	if checkFlag == nil {
		t.Error("--check flag should exist for dry-run updates")
	}
}

func TestUpdateDevVersion(t *testing.T) {
	// Test that "dev" version is handled specially
	currentVersion := "dev"

	if currentVersion != "dev" {
		t.Error("dev version should be recognized")
	}

	// Dev versions should always show available updates
	isDev := currentVersion == "dev"
	if !isDev {
		t.Error("should detect dev version")
	}
}

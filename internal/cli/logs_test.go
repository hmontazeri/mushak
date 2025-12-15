package cli

import (
	"strings"
	"testing"
)

func TestLogsCommand(t *testing.T) {
	if logsCmd == nil {
		t.Fatal("logsCmd should not be nil")
	}

	if logsCmd.Use != "logs" {
		t.Errorf("logsCmd.Use = %v, want logs", logsCmd.Use)
	}

	if logsCmd.Short == "" {
		t.Error("logsCmd.Short should not be empty")
	}

	if logsCmd.Long == "" {
		t.Error("logsCmd.Long should not be empty")
	}

	if logsCmd.RunE == nil {
		t.Error("logsCmd.RunE should not be nil")
	}
}

func TestLogsCommandFlags(t *testing.T) {
	if logsCmd == nil {
		t.Fatal("logsCmd should not be nil")
	}

	// Check tail flag
	tailFlag := logsCmd.Flags().Lookup("tail")
	if tailFlag == nil {
		t.Error("logs command should have --tail flag")
	} else {
		if tailFlag.Shorthand != "n" {
			t.Errorf("tail flag shorthand = %v, want n", tailFlag.Shorthand)
		}
		if tailFlag.DefValue != "100" {
			t.Errorf("tail flag default = %v, want 100", tailFlag.DefValue)
		}
	}

	// Check follow flag
	followFlag := logsCmd.Flags().Lookup("follow")
	if followFlag == nil {
		t.Error("logs command should have --follow flag")
	} else {
		if followFlag.Shorthand != "f" {
			t.Errorf("follow flag shorthand = %v, want f", followFlag.Shorthand)
		}
		if followFlag.DefValue != "true" {
			t.Errorf("follow flag default = %v, want true", followFlag.DefValue)
		}
	}

	// Check key flag
	keyFlag := logsCmd.Flags().Lookup("key")
	if keyFlag == nil {
		t.Error("logs command should have --key flag")
	}
}

func TestLogsCommandDescription(t *testing.T) {
	if logsCmd == nil {
		t.Fatal("logsCmd should not be nil")
	}

	requiredDescriptions := []string{
		"streams",
		"container",
	}

	for _, desc := range requiredDescriptions {
		if !strings.Contains(logsCmd.Long, desc) && !strings.Contains(logsCmd.Short, desc) {
			t.Errorf("logs command description should mention %q", desc)
		}
	}
}

package cli

import (
	"strings"
	"testing"
)

func TestContainersCommand(t *testing.T) {
	if containersCmd == nil {
		t.Fatal("containersCmd should not be nil")
	}

	if containersCmd.Use != "containers" {
		t.Errorf("containersCmd.Use = %v, want containers", containersCmd.Use)
	}

	if containersCmd.Short == "" {
		t.Error("containersCmd.Short should not be empty")
	}

	if containersCmd.RunE == nil {
		t.Error("containersCmd.RunE should not be nil")
	}

	requiredDescriptions := []string{
		"container",
	}

	for _, desc := range requiredDescriptions {
		if !strings.Contains(containersCmd.Long, desc) && !strings.Contains(containersCmd.Short, desc) {
			t.Errorf("containers command description should mention %q", desc)
		}
	}
}

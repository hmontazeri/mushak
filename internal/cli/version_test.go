package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestVersionCommand(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	oldColorOutput := color.Output
	r, w, _ := os.Pipe()
	os.Stdout = w
	color.Output = w

	// Restore output deferred
	defer func() {
		os.Stdout = oldStdout
		color.Output = oldColorOutput
	}()

	// Execute command
	versionCmd.Run(versionCmd, []string{})

	// Close write end to read
	w.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	output := buf.String()

	// ui.PrintInfo adds "→ " prefix
	if !strings.Contains(output, "mushak version") {
		t.Errorf("version command output should contain 'mushak version', got: %s", output)
	}
}

func TestVersionCommandStructure(t *testing.T) {
	if versionCmd == nil {
		t.Fatal("versionCmd should not be nil")
	}

	if versionCmd.Use != "version" {
		t.Errorf("versionCmd.Use = %v, want version", versionCmd.Use)
	}

	if versionCmd.Short == "" {
		t.Error("versionCmd.Short should not be empty")
	}

	if versionCmd.Run == nil {
		t.Error("versionCmd.Run should not be nil")
	}
}

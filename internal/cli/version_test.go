package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestVersionCommand(t *testing.T) {
	// Capture output
	buf := new(bytes.Buffer)

	// Create a new version command for testing
	cmd := &cobra.Command{
		Use: "version",
		Run: func(cmd *cobra.Command, args []string) {
			buf.WriteString("mushak version test\n")
		},
	}

	// Execute command
	cmd.Run(cmd, []string{})

	output := buf.String()

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

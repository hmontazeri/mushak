package cli

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func TestWithTimer(t *testing.T) {
	// Capture stdout
	oldStdout := color.Output
	r, w, _ := os.Pipe()
	color.Output = w

	defer func() {
		color.Output = oldStdout
	}()

	t.Run("success", func(t *testing.T) {
		// Create a dummy command
		cmd := &cobra.Command{
			Use: "test-cmd",
			RunE: func(cmd *cobra.Command, args []string) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
		}

		// Wrap with timer
		wrapped := withTimer(cmd.RunE)

		// Execute
		err := wrapped(cmd, []string{})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Read output
		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		// Verify output
		if !strings.Contains(output, "test-cmd completed in") {
			t.Errorf("expected output to contain timing info, got: %q", output)
		}
	})

	t.Run("failure", func(t *testing.T) {
		// Reset pipe for new test
		r, w, _ = os.Pipe()
		color.Output = w
		
		cmd := &cobra.Command{
			Use: "fail-cmd",
			RunE: func(cmd *cobra.Command, args []string) error {
				return errors.New("failed")
			},
		}

		wrapped := withTimer(cmd.RunE)
		err := wrapped(cmd, []string{})
		
		if err == nil {
			t.Error("expected error")
		}

		w.Close()
		var buf bytes.Buffer
		io.Copy(&buf, r)
		output := buf.String()

		if output != "" {
			t.Errorf("expected no output on error, got: %q", output)
		}
	})
}

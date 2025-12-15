package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

// Executor handles remote command execution
type Executor struct {
	client *Client
}

// NewExecutor creates a new executor for the given client
func NewExecutor(client *Client) *Executor {
	return &Executor{client: client}
}

// Run executes a command and returns stdout
func (e *Executor) Run(cmd string) (string, error) {
	return e.RunWithContext(context.Background(), cmd)
}

// RunWithContext executes a command with a context
func (e *Executor) RunWithContext(ctx context.Context, cmd string) (string, error) {
	session, err := e.client.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Run command in a goroutine to support context cancellation
	errChan := make(chan error, 1)
	go func() {
		errChan <- session.Run(cmd)
	}()

	select {
	case <-ctx.Done():
		session.Signal(ssh.SIGKILL)
		return "", ctx.Err()
	case err := <-errChan:
		if err != nil {
			return "", fmt.Errorf("command failed: %w\nstderr: %s", err, stderr.String())
		}
	}

	return stdout.String(), nil
}

// RunWithTimeout executes a command with a timeout
func (e *Executor) RunWithTimeout(cmd string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return e.RunWithContext(ctx, cmd)
}

// StreamRun executes a command and streams output to provided writers
func (e *Executor) StreamRun(cmd string, stdout, stderr io.Writer) error {
	session, err := e.client.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	if err := session.Run(cmd); err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

// RunSudo executes a command with sudo
func (e *Executor) RunSudo(cmd string) (string, error) {
	return e.Run("sudo " + cmd)
}

// FileExists checks if a file exists on the remote server
func (e *Executor) FileExists(path string) (bool, error) {
	_, err := e.Run(fmt.Sprintf("test -f %s", path))
	if err != nil {
		// test command returns error if file doesn't exist
		return false, nil
	}
	return true, nil
}

// DirExists checks if a directory exists on the remote server
func (e *Executor) DirExists(path string) (bool, error) {
	_, err := e.Run(fmt.Sprintf("test -d %s", path))
	if err != nil {
		return false, nil
	}
	return true, nil
}

// WriteFile writes content to a file on the remote server
func (e *Executor) WriteFile(path, content string) error {
	// Use heredoc to write file content
	cmd := fmt.Sprintf("cat > %s << 'MUSHAK_EOF'\n%s\nMUSHAK_EOF", path, content)
	_, err := e.Run(cmd)
	return err
}

// WriteFileSudo writes content to a file with sudo
func (e *Executor) WriteFileSudo(path, content string) error {
	cmd := fmt.Sprintf("sudo tee %s > /dev/null << 'MUSHAK_EOF'\n%s\nMUSHAK_EOF", path, content)
	_, err := e.Run(cmd)
	return err
}

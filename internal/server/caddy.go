package server

import (
	"fmt"

	"github.com/hmontazeri/mushak/internal/ssh"
)

// InitializeCaddyMultiApp sets up Caddy for multi-app support
func InitializeCaddyMultiApp(executor *ssh.Executor) error {
	fmt.Println("Initializing Caddy multi-app setup...")

	// Create apps directory
	if _, err := executor.RunSudo("mkdir -p /etc/caddy/apps"); err != nil {
		return fmt.Errorf("failed to create Caddy apps directory: %w", err)
	}

	// Check if main Caddyfile already has the import statement
	exists, _ := executor.FileExists("/etc/caddy/Caddyfile")

	mainCaddyfile := `# Mushak multi-app Caddyfile
# Import all app configurations
import /etc/caddy/apps/*.caddy
`

	if !exists {
		// Create new Caddyfile with import statement
		if err := executor.WriteFileSudo("/etc/caddy/Caddyfile", mainCaddyfile); err != nil {
			return fmt.Errorf("failed to create main Caddyfile: %w", err)
		}
	} else {
		// Check if import already exists
		content, err := executor.Run("cat /etc/caddy/Caddyfile")
		if err != nil {
			return fmt.Errorf("failed to read Caddyfile: %w", err)
		}

		// If import statement doesn't exist, add it
		if len(content) > 0 && content != mainCaddyfile {
			// Backup existing Caddyfile
			if _, err := executor.RunSudo("cp /etc/caddy/Caddyfile /etc/caddy/Caddyfile.backup"); err != nil {
				return fmt.Errorf("failed to backup Caddyfile: %w", err)
			}

			// Write new Caddyfile
			if err := executor.WriteFileSudo("/etc/caddy/Caddyfile", mainCaddyfile); err != nil {
				return fmt.Errorf("failed to update Caddyfile: %w", err)
			}

			fmt.Println("⚠ Existing Caddyfile backed up to /etc/caddy/Caddyfile.backup")
		}
	}

	fmt.Println("✓ Caddy multi-app setup initialized")
	return nil
}

// CreateAppCaddyConfig creates or updates the Caddy config for an app
func CreateAppCaddyConfig(executor *ssh.Executor, appName, domain string, port int) error {
	fmt.Printf("Updating Caddy config for %s...\n", appName)

	configPath := fmt.Sprintf("/etc/caddy/apps/%s.caddy", appName)

	config := fmt.Sprintf(`%s {
	reverse_proxy localhost:%d
}
`, domain, port)

	if err := executor.WriteFileSudo(configPath, config); err != nil {
		return fmt.Errorf("failed to write Caddy config: %w", err)
	}

	// Reload Caddy
	if err := ReloadCaddy(executor); err != nil {
		return err
	}

	fmt.Println("✓ Caddy configuration updated")
	return nil
}

// RemoveAppCaddyConfig removes the Caddy config for an app
func RemoveAppCaddyConfig(executor *ssh.Executor, appName string) error {
	fmt.Printf("Removing Caddy config for %s...\n", appName)

	configPath := fmt.Sprintf("/etc/caddy/apps/%s.caddy", appName)

	// Remove config file
	if _, err := executor.RunSudo(fmt.Sprintf("rm -f %s", configPath)); err != nil {
		return fmt.Errorf("failed to remove Caddy config: %w", err)
	}

	// Reload Caddy
	if err := ReloadCaddy(executor); err != nil {
		return err
	}

	fmt.Println("✓ Caddy configuration removed")
	return nil
}

// ReloadCaddy reloads the Caddy server
func ReloadCaddy(executor *ssh.Executor) error {
	// Try systemctl reload first
	if _, err := executor.RunSudo("systemctl reload caddy"); err != nil {
		// Fall back to caddy reload command
		if _, err := executor.RunSudo("caddy reload --config /etc/caddy/Caddyfile"); err != nil {
			return fmt.Errorf("failed to reload Caddy: %w", err)
		}
	}

	return nil
}

package server

import (
	"fmt"
	"strings"

	"github.com/hmontazeri/mushak/internal/ssh"
	"github.com/hmontazeri/mushak/internal/ui"
)

// InstallDependencies installs all required dependencies on the server
func InstallDependencies(executor *ssh.Executor) error {
	ui.PrintInfo("Installing dependencies...")

	if err := InstallGit(executor); err != nil {
		return err
	}

	if err := InstallDocker(executor); err != nil {
		return err
	}

	if err := InstallCaddy(executor); err != nil {
		return err
	}

	ui.PrintSuccess("All dependencies installed successfully")
	return nil
}

// InstallGit installs Git if not already present
func InstallGit(executor *ssh.Executor) error {
	ui.PrintInfo("Checking Git...")

	// Check if git is installed
	if _, err := executor.Run("which git"); err == nil {
		ui.PrintSuccess("Git already installed")
		return nil
	}

	ui.PrintInfo("Installing Git...")

	// Update package list
	if _, err := executor.RunSudo("apt-get update -qq"); err != nil {
		return fmt.Errorf("failed to update package list: %w", err)
	}

	// Install git
	if _, err := executor.RunSudo("DEBIAN_FRONTEND=noninteractive apt-get install -y git"); err != nil {
		return fmt.Errorf("failed to install git: %w", err)
	}

	ui.PrintSuccess("Git installed")
	return nil
}

// InstallDocker installs Docker if not already present
func InstallDocker(executor *ssh.Executor) error {
	ui.PrintInfo("Checking Docker...")

	// Check if docker is installed
	if _, err := executor.Run("which docker"); err == nil {
		ui.PrintSuccess("Docker already installed")
		return nil
	}

	ui.PrintInfo("Installing Docker...")

	// Install prerequisites
	prereqs := []string{
		"apt-transport-https",
		"ca-certificates",
		"curl",
		"gnupg",
		"lsb-release",
	}

	if _, err := executor.RunSudo(fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", strings.Join(prereqs, " "))); err != nil {
		return fmt.Errorf("failed to install prerequisites: %w", err)
	}

	// Add Docker's official GPG key
	commands := []string{
		"mkdir -p /etc/apt/keyrings",
		"curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg",
		"chmod a+r /etc/apt/keyrings/docker.gpg",
	}

	for _, cmd := range commands {
		if _, err := executor.RunSudo(cmd); err != nil {
			return fmt.Errorf("failed to add Docker GPG key: %w", err)
		}
	}

	// Set up the repository
	repoCmd := `echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null`
	if _, err := executor.RunSudo(repoCmd); err != nil {
		return fmt.Errorf("failed to add Docker repository: %w", err)
	}

	// Update package list
	if _, err := executor.RunSudo("apt-get update -qq"); err != nil {
		return fmt.Errorf("failed to update package list: %w", err)
	}

	// Install Docker Engine
	if _, err := executor.RunSudo("DEBIAN_FRONTEND=noninteractive apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin"); err != nil {
		return fmt.Errorf("failed to install Docker: %w", err)
	}

	// Start Docker service
	if _, err := executor.RunSudo("systemctl start docker"); err != nil {
		return fmt.Errorf("failed to start Docker: %w", err)
	}

	if _, err := executor.RunSudo("systemctl enable docker"); err != nil {
		return fmt.Errorf("failed to enable Docker: %w", err)
	}

	// Configure Docker BuildKit GC for better cache management
	if err := ConfigureDockerGC(executor); err != nil {
		ui.PrintWarning(fmt.Sprintf("Could not configure Docker GC: %v", err))
	}

	// Add user to docker group
	if _, err := executor.RunSudo("usermod -aG docker $USER"); err != nil {
		// Not critical, user can run docker with sudo
		ui.PrintWarning("Could not add user to docker group")
	}

	ui.PrintSuccess("Docker installed")
	return nil
}

// ConfigureDockerGC sets up sensible default garbage collection for Docker BuildKit
func ConfigureDockerGC(executor *ssh.Executor) error {
	ui.PrintInfo("Configuring Docker BuildKit GC...")

	daemonJsonPath := "/etc/docker/daemon.json"
	
	// Check if daemon.json already exists
	exists := false
	if _, err := executor.Run(fmt.Sprintf("ls %s", daemonJsonPath)); err == nil {
		exists = true
	}

	gcConfig := `{
  "builder": {
    "gc": {
      "enabled": true,
      "default_keep_storage": "10GB",
      "policy": [
        { "keep_storage": "5GB", "filter": ["unused-for=168h"] },
        { "keep_storage": "10GB", "filter": ["unused-for=336h"] },
        { "keep_storage": "20GB" }
      ]
    }
  }
}`

	if exists {
		// If it exists, we don't want to blindly overwrite it. 
		// For now, we'll just check if "builder" is already there.
		content, err := executor.RunSudo(fmt.Sprintf("cat %s", daemonJsonPath))
		if err == nil && strings.Contains(content, "\"builder\"") {
			ui.PrintInfo("Docker builder configuration already exists, skipping...")
			return nil
		}
		
		// If it exists but no builder config, we'd need to merge JSON which is complex via SSH.
		// For simplicity in Mushak, we'll just inform the user if we can't easily merge.
		ui.PrintWarning("Custom /etc/docker/daemon.json exists. Please manually add 'builder' GC config.")
		return nil
	}

	// Create new daemon.json
	if err := executor.WriteFileSudo(daemonJsonPath, gcConfig); err != nil {
		return fmt.Errorf("failed to write daemon.json: %w", err)
	}

	// Reload docker to apply changes
	if _, err := executor.RunSudo("systemctl reload docker"); err != nil {
		// Some systems might not support reload for daemon.json changes, try restart
		if _, err := executor.RunSudo("systemctl restart docker"); err != nil {
			return fmt.Errorf("failed to restart docker after config change: %w", err)
		}
	}

	ui.PrintSuccess("Docker BuildKit GC configured (10GB limit)")
	return nil
}

// InstallCaddy installs Caddy if not already present
func InstallCaddy(executor *ssh.Executor) error {
	ui.PrintInfo("Checking Caddy...")

	// Check if caddy is installed
	if _, err := executor.Run("which caddy"); err == nil {
		ui.PrintSuccess("Caddy already installed")
		return nil
	}

	ui.PrintInfo("Installing Caddy...")

	// Install prerequisites
	if _, err := executor.RunSudo("DEBIAN_FRONTEND=noninteractive apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl"); err != nil {
		return fmt.Errorf("failed to install prerequisites: %w", err)
	}

	// Add Caddy's official GPG key and repository
	commands := []string{
		"curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg",
		"curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list",
	}

	for _, cmd := range commands {
		if _, err := executor.RunSudo(cmd); err != nil {
			return fmt.Errorf("failed to add Caddy repository: %w", err)
		}
	}

	// Update package list
	if _, err := executor.RunSudo("apt-get update -qq"); err != nil {
		return fmt.Errorf("failed to update package list: %w", err)
	}

	// Install Caddy
	if _, err := executor.RunSudo("DEBIAN_FRONTEND=noninteractive apt-get install -y caddy"); err != nil {
		return fmt.Errorf("failed to install Caddy: %w", err)
	}

	// Start Caddy service
	if _, err := executor.RunSudo("systemctl start caddy"); err != nil {
		return fmt.Errorf("failed to start Caddy: %w", err)
	}

	if _, err := executor.RunSudo("systemctl enable caddy"); err != nil {
		return fmt.Errorf("failed to enable Caddy: %w", err)
	}

	ui.PrintSuccess("Caddy installed")
	return nil
}

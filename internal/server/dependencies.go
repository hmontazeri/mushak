package server

import (
	"fmt"
	"strings"

	"github.com/hmontazeri/mushak/internal/ssh"
)

// InstallDependencies installs all required dependencies on the server
func InstallDependencies(executor *ssh.Executor) error {
	fmt.Println("Installing dependencies...")

	if err := InstallGit(executor); err != nil {
		return err
	}

	if err := InstallDocker(executor); err != nil {
		return err
	}

	if err := InstallCaddy(executor); err != nil {
		return err
	}

	fmt.Println("✓ All dependencies installed successfully")
	return nil
}

// InstallGit installs Git if not already present
func InstallGit(executor *ssh.Executor) error {
	fmt.Print("Checking Git... ")

	// Check if git is installed
	if _, err := executor.Run("which git"); err == nil {
		fmt.Println("already installed")
		return nil
	}

	fmt.Println("installing...")

	// Update package list
	if _, err := executor.RunSudo("apt-get update -qq"); err != nil {
		return fmt.Errorf("failed to update package list: %w", err)
	}

	// Install git
	if _, err := executor.RunSudo("DEBIAN_FRONTEND=noninteractive apt-get install -y git"); err != nil {
		return fmt.Errorf("failed to install git: %w", err)
	}

	fmt.Println("✓ Git installed")
	return nil
}

// InstallDocker installs Docker if not already present
func InstallDocker(executor *ssh.Executor) error {
	fmt.Print("Checking Docker... ")

	// Check if docker is installed
	if _, err := executor.Run("which docker"); err == nil {
		fmt.Println("already installed")
		return nil
	}

	fmt.Println("installing...")

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

	// Add user to docker group
	if _, err := executor.RunSudo("usermod -aG docker $USER"); err != nil {
		// Not critical, user can run docker with sudo
		fmt.Println("⚠ Warning: Could not add user to docker group")
	}

	fmt.Println("✓ Docker installed")
	return nil
}

// InstallCaddy installs Caddy if not already present
func InstallCaddy(executor *ssh.Executor) error {
	fmt.Print("Checking Caddy... ")

	// Check if caddy is installed
	if _, err := executor.Run("which caddy"); err == nil {
		fmt.Println("already installed")
		return nil
	}

	fmt.Println("installing...")

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

	fmt.Println("✓ Caddy installed")
	return nil
}

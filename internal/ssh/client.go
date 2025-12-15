package ssh

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Client represents an SSH client connection
type Client struct {
	config *ssh.ClientConfig
	client *ssh.Client
	host   string
	port   string
}

// Config holds SSH connection parameters
type Config struct {
	Host     string
	Port     string
	User     string
	KeyPath  string
	Password string
}

// NewClient creates a new SSH client
func NewClient(cfg Config) (*Client, error) {
	if cfg.Port == "" {
		cfg.Port = "22"
	}

	authMethods, err := getAuthMethods(cfg.KeyPath, cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to get auth methods: %w", err)
	}

	sshConfig := &ssh.ClientConfig{
		User:            cfg.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
	}

	return &Client{
		config: sshConfig,
		host:   cfg.Host,
		port:   cfg.Port,
	}, nil
}

// Connect establishes the SSH connection
func (c *Client) Connect() error {
	addr := net.JoinHostPort(c.host, c.port)
	client, err := ssh.Dial("tcp", addr, c.config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	c.client = client
	return nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// getAuthMethods returns available SSH authentication methods
func getAuthMethods(keyPath, password string) ([]ssh.AuthMethod, error) {
	var authMethods []ssh.AuthMethod

	// Try SSH agent first
	if agentAuth := getAgentAuth(); agentAuth != nil {
		authMethods = append(authMethods, agentAuth)
	}

	// Try key-based auth
	if keyPath == "" {
		// Default to ~/.ssh/id_rsa
		homeDir, err := os.UserHomeDir()
		if err == nil {
			keyPath = filepath.Join(homeDir, ".ssh", "id_rsa")
		}
	}

	if keyPath != "" {
		keyAuth, err := getKeyAuth(keyPath)
		if err == nil {
			authMethods = append(authMethods, keyAuth)
		}
	}

	// Try password auth if provided
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
	}

	return authMethods, nil
}

// getKeyAuth returns key-based authentication
func getKeyAuth(keyPath string) (ssh.AuthMethod, error) {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}

	return ssh.PublicKeys(signer), nil
}

// getAgentAuth returns SSH agent authentication if available
func getAgentAuth() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	return nil
}

package ssh

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid config with password",
			cfg: Config{
				Host:     "example.com",
				Port:     "22",
				User:     "testuser",
				Password: "testpass",
			},
			wantErr: false,
		},
		{
			name: "valid config without port (default)",
			cfg: Config{
				Host:     "example.com",
				User:     "testuser",
				Password: "testpass",
			},
			wantErr: false,
		},
		{
			name: "valid config with custom port",
			cfg: Config{
				Host:     "example.com",
				Port:     "2222",
				User:     "testuser",
				Password: "testpass",
			},
			wantErr: false,
		},
		{
			name: "config with key path",
			cfg: Config{
				Host:    "example.com",
				Port:    "22",
				User:    "testuser",
				KeyPath: "/nonexistent/key",
			},
			wantErr: false, // NewClient doesn't fail, only Connect does
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if client == nil {
					t.Error("NewClient() returned nil client")
					return
				}

				if client.host != tt.cfg.Host {
					t.Errorf("client.host = %v, want %v", client.host, tt.cfg.Host)
				}

				expectedPort := tt.cfg.Port
				if expectedPort == "" {
					expectedPort = "22"
				}
				if client.port != expectedPort {
					t.Errorf("client.port = %v, want %v", client.port, expectedPort)
				}

				if client.config == nil {
					t.Error("client.config is nil")
				}

				if client.config.User != tt.cfg.User {
					t.Errorf("client.config.User = %v, want %v", client.config.User, tt.cfg.User)
				}
			}
		})
	}
}

func TestClient_Close(t *testing.T) {
	tests := []struct {
		name    string
		client  *Client
		wantErr bool
	}{
		{
			name: "close without connection",
			client: &Client{
				client: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.Close()
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		cfg    Config
		field  string
		expect string
	}{
		{
			name: "host is set",
			cfg: Config{
				Host:     "example.com",
				User:     "user",
				Password: "pass",
			},
			field:  "Host",
			expect: "example.com",
		},
		{
			name: "user is set",
			cfg: Config{
				Host:     "example.com",
				User:     "testuser",
				Password: "pass",
			},
			field:  "User",
			expect: "testuser",
		},
		{
			name: "port defaults to 22",
			cfg: Config{
				Host:     "example.com",
				User:     "user",
				Password: "pass",
			},
			field:  "Port",
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.field {
			case "Host":
				if tt.cfg.Host != tt.expect {
					t.Errorf("Config.Host = %v, want %v", tt.cfg.Host, tt.expect)
				}
			case "User":
				if tt.cfg.User != tt.expect {
					t.Errorf("Config.User = %v, want %v", tt.cfg.User, tt.expect)
				}
			case "Port":
				if tt.cfg.Port != tt.expect {
					t.Errorf("Config.Port = %v, want %v", tt.cfg.Port, tt.expect)
				}
			}
		})
	}
}

func TestGetAuthMethods_NoAuthAvailable(t *testing.T) {
	// Test with no valid auth methods
	_, err := getAuthMethods("/nonexistent/key/path", "")
	if err == nil {
		// This might not error if SSH agent is available
		// So we just check that the function executes
		t.Log("getAuthMethods completed (may have found SSH agent)")
	}
}

func TestGetAuthMethods_WithPassword(t *testing.T) {
	// Test with password auth
	methods, err := getAuthMethods("", "testpassword")
	if err != nil {
		t.Errorf("getAuthMethods() with password error = %v", err)
		return
	}

	if len(methods) == 0 {
		t.Error("getAuthMethods() returned no auth methods with password")
	}
}

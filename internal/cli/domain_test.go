package cli

import (
	"testing"
)

func TestGetPortFromCaddyConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		want    int
		wantErr bool
	}{
		{
			name: "valid config",
			config: `example.com {
	reverse_proxy localhost:8080
}`,
			want:    8080,
			wantErr: false,
		},
		{
			name: "valid config with different spacing",
			config: `example.com {
    reverse_proxy   localhost:9000
}`,
			want:    9000,
			wantErr: false,
		},
		{
			name: "config with comments",
			config: `# This is a comment
example.com {
	reverse_proxy localhost:3000 # App port
}`,
			want:    3000,
			wantErr: false,
		},
		{
			name:    "invalid config - no port",
			config:  `example.com { file_server }`,
			want:    0,
			wantErr: true,
		},
		{
			name: "invalid port number",
			config: `example.com {
	reverse_proxy localhost:abc
}`,
			want:    0,
			wantErr: true, // Regex won't match, so it returns "could not find port" error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPortFromCaddyConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPortFromCaddyConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getPortFromCaddyConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

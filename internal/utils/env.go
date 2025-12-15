package utils

import (
	"fmt"
	"os"
	"strings"
)

// DetectLocalEnvFile finds the local env file with priority: .env.prod > .env.production > .env
func DetectLocalEnvFile() (string, error) {
	candidates := []string{".env.prod", ".env.production", ".env"}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("no environment file found (tried: %s)", strings.Join(candidates, ", "))
}

// ParseEnvFile parses a .env file and returns a map of key-value pairs
func ParseEnvFile(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	vars := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			vars[key] = value
		}
	}

	return vars, nil
}

// CountEnvVars returns the number of variables in an env file
func CountEnvVars(path string) (int, error) {
	vars, err := ParseEnvFile(path)
	if err != nil {
		return 0, err
	}
	return len(vars), nil
}

// GetEnvVarKeys returns the keys (variable names) from an env file
func GetEnvVarKeys(path string) ([]string, error) {
	vars, err := ParseEnvFile(path)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	return keys, nil
}

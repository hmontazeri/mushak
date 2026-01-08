package utils

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// DetectInternalPort attempts to find the port the application listens on
func DetectInternalPort() int {
	// 1. Check Dockerfile
	if port := detectFromDockerfile(); port > 0 {
		return port
	}

	// 2. Check docker-compose.yml
	if port := detectFromCompose(); port > 0 {
		return port
	}

	return 0
}

func detectFromDockerfile() int {
	f, err := os.Open("Dockerfile")
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	re := regexp.MustCompile(`(?i)^EXPOSE\s+(\d+)`)
	for scanner.Scan() {
		matches := re.FindStringSubmatch(scanner.Text())
		if len(matches) > 1 {
			port, _ := strconv.Atoi(matches[1])
			return port
		}
	}
	return 0
}

func detectFromCompose() int {
	filename := ""
	if _, err := os.Stat("docker-compose.yml"); err == nil {
		filename = "docker-compose.yml"
	} else if _, err := os.Stat("docker-compose.yaml"); err == nil {
		filename = "docker-compose.yaml"
	}

	if filename == "" {
		return 0
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}

	var config struct {
		Services map[string]struct {
			Ports []string `yaml:"ports"`
		} `yaml:"services"`
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return 0
	}

	for _, svc := range config.Services {
		for _, p := range svc.Ports {
			// Handle various formats: "80:80", "80", "127.0.0.1:80:80"
			parts := strings.Split(p, ":")
			last := parts[len(parts)-1]
			if port, err := strconv.Atoi(last); err == nil {
				return port
			}
		}
	}

	return 0
}

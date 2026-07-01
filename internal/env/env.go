package env

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func init() {
	LoadEnv()
}

// LoadEnv reads .env files from the current working directory, and/or ~/.codeforge/.env,
// setting environment variables if they are not already set.
func LoadEnv() {
	// 1. Try to load from current working directory
	_ = loadEnvFile(".env")

	// 2. Try to load from ~/.codeforge/.env
	if home, err := os.UserHomeDir(); err == nil {
		_ = loadEnvFile(filepath.Join(home, ".codeforge", ".env"))
	}
}

func loadEnvFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		// Strip surrounding quotes
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
		}

		// Only set if not already present in environment
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return nil
}

// GetAPIURL builds the API URL using CODEFORGE_API_URL or CODEFORGE_HOST and CODEFORGE_PORT.
func GetAPIURL(defaultPort int) string {
	if apiURL := os.Getenv("CODEFORGE_API_URL"); apiURL != "" {
		return strings.TrimSuffix(apiURL, "/")
	}

	host := os.Getenv("CODEFORGE_HOST")
	if host == "" {
		host = "localhost"
	}

	protocol := "http://"
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		protocol = ""
	}

	portStr := os.Getenv("CODEFORGE_PORT")
	if portStr == "" {
		hostName := strings.TrimPrefix(strings.TrimPrefix(host, "https://"), "http://")
		if hostName == "localhost" || hostName == "127.0.0.1" {
			portStr = strconv.Itoa(defaultPort)
		}
	}

	if portStr != "" && portStr != "none" {
		hostName := strings.TrimPrefix(strings.TrimPrefix(host, "https://"), "http://")
		if !strings.Contains(hostName, ":") {
			return fmt.Sprintf("%s%s:%s", protocol, host, portStr)
		}
	}

	return fmt.Sprintf("%s%s", protocol, host)
}

// GetPort reads CODEFORGE_PORT or returns the default port.
func GetPort(defaultPort int) int {
	if portStr := os.Getenv("CODEFORGE_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			return p
		}
	}
	return defaultPort
}

package env

import (
	"os"
	"testing"
)

func TestGetAPIURL(t *testing.T) {
	// Clean up environment variables at the end
	defer func() {
		os.Unsetenv("CODEFORGE_API_URL")
		os.Unsetenv("CODEFORGE_HOST")
		os.Unsetenv("CODEFORGE_PORT")
	}()

	tests := []struct {
		name        string
		apiURL      string
		host        string
		port        string
		defaultPort int
		expected    string
	}{
		{
			name:        "Defaults to localhost",
			apiURL:      "",
			host:        "",
			port:        "",
			defaultPort: 7080,
			expected:    "http://localhost:7080",
		},
		{
			name:        "Uses CODEFORGE_API_URL explicitly",
			apiURL:      "https://apicodeforge.khajumsanjog.com",
			host:        "",
			port:        "",
			defaultPort: 7080,
			expected:    "https://apicodeforge.khajumsanjog.com",
		},
		{
			name:        "Uses CODEFORGE_HOST and CODEFORGE_PORT",
			apiURL:      "",
			host:        "apicodeforge.khajumsanjog.com",
			port:        "8443",
			defaultPort: 7080,
			expected:    "http://apicodeforge.khajumsanjog.com:8443",
		},
		{
			name:        "Uses CODEFORGE_HOST without port on production domains",
			apiURL:      "",
			host:        "apicodeforge.khajumsanjog.com",
			port:        "",
			defaultPort: 7080,
			expected:    "http://apicodeforge.khajumsanjog.com",
		},
		{
			name:        "Host with protocol",
			apiURL:      "",
			host:        "https://my-api-host.com",
			port:        "",
			defaultPort: 7080,
			expected:    "https://my-api-host.com",
		},
		{
			name:        "Host with protocol and custom port",
			apiURL:      "",
			host:        "https://my-api-host.com",
			port:        "9000",
			defaultPort: 7080,
			expected:    "https://my-api-host.com:9000",
		},
		{
			name:        "Localhost with custom port in env",
			apiURL:      "",
			host:        "localhost",
			port:        "7090",
			defaultPort: 7080,
			expected:    "http://localhost:7090",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("CODEFORGE_API_URL")
			os.Unsetenv("CODEFORGE_HOST")
			os.Unsetenv("CODEFORGE_PORT")

			if tt.apiURL != "" {
				os.Setenv("CODEFORGE_API_URL", tt.apiURL)
			}
			if tt.host != "" {
				os.Setenv("CODEFORGE_HOST", tt.host)
			}
			if tt.port != "" {
				os.Setenv("CODEFORGE_PORT", tt.port)
			}

			result := GetAPIURL(tt.defaultPort)
			if result != tt.expected {
				t.Errorf("GetAPIURL() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestLoadEnvFile(t *testing.T) {
	// Create a temp .env file
	tmpFile, err := os.CreateTemp("", "test-.env")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `
# This is a comment
TEST_KEY_1=value1
TEST_KEY_2="value2"
TEST_KEY_3='value3'
`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	os.Unsetenv("TEST_KEY_1")
	os.Unsetenv("TEST_KEY_2")
	os.Unsetenv("TEST_KEY_3")

	err = loadEnvFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("loadEnvFile failed: %v", err)
	}

	if val := os.Getenv("TEST_KEY_1"); val != "value1" {
		t.Errorf("Expected TEST_KEY_1 to be 'value1', got %q", val)
	}
	if val := os.Getenv("TEST_KEY_2"); val != "value2" {
		t.Errorf("Expected TEST_KEY_2 to be 'value2', got %q", val)
	}
	if val := os.Getenv("TEST_KEY_3"); val != "value3" {
		t.Errorf("Expected TEST_KEY_3 to be 'value3', got %q", val)
	}

	// Clean up
	os.Unsetenv("TEST_KEY_1")
	os.Unsetenv("TEST_KEY_2")
	os.Unsetenv("TEST_KEY_3")
}

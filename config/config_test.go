package config

import (
	"testing"
	"time"
)

// MockEnvGetter implements EnvGetter for testing
type MockEnvGetter struct {
	envVars map[string]string
}

func NewMockEnvGetter(envVars map[string]string) *MockEnvGetter {
	return &MockEnvGetter{envVars: envVars}
}

func (m *MockEnvGetter) Getenv(key string) string {
	return m.envVars[key]
}

type NewConfigTestCase struct {
	name        string
	envVars     map[string]string
	expectError bool
	expected    *Config
}

func TestNewConfig(t *testing.T) {
	testCases := []NewConfigTestCase{
		{
			name: "All required variables present",
			envVars: map[string]string{
				"STARFEED_GITHUB_API_TOKEN":   "gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: false,
			expected: &Config{
				GithubToken:   "gh_token123",
				FreshRssUrl:   "http://freshrss.example.com",
				FreshRssUser:  "testuser",
				FreshRssToken: "freshrss_token456",
				DebugMode:     false,
				SingleRunMode: false,
				HttpTimeout:   10 * time.Second,
			},
		},
		{
			name: "All variables present with debug and single run mode",
			envVars: map[string]string{
				"STARFEED_GITHUB_API_TOKEN":   "gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
				"STARFEED_DEBUG_MODE":         "true",
				"STARFEED_SINGLE_RUN_MODE":    "true",
				"STARFEED_HTTP_TIMEOUT":       "30",
			},
			expectError: false,
			expected: &Config{
				GithubToken:   "gh_token123",
				FreshRssUrl:   "http://freshrss.example.com",
				FreshRssUser:  "testuser",
				FreshRssToken: "freshrss_token456",
				DebugMode:     true,
				SingleRunMode: true,
				HttpTimeout:   30 * time.Second,
			},
		},
		{
			name: "Missing GitHub token should error",
			envVars: map[string]string{
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing FreshRSS URL should error",
			envVars: map[string]string{
				"STARFEED_GITHUB_API_TOKEN":   "gh_token123",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing FreshRSS user should error",
			envVars: map[string]string{
				"STARFEED_GITHUB_API_TOKEN":   "gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing FreshRSS token should error",
			envVars: map[string]string{
				"STARFEED_GITHUB_API_TOKEN": "gh_token123",
				"STARFEED_FRESHRSS_URL":     "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":    "testuser",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid HTTP timeout defaults to 10 seconds",
			envVars: map[string]string{
				"STARFEED_GITHUB_API_TOKEN":   "gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
				"STARFEED_HTTP_TIMEOUT":       "invalid",
			},
			expectError: false,
			expected: &Config{
				GithubToken:   "gh_token123",
				FreshRssUrl:   "http://freshrss.example.com",
				FreshRssUser:  "testuser",
				FreshRssToken: "freshrss_token456",
				DebugMode:     false,
				SingleRunMode: false,
				HttpTimeout:   10 * time.Second,
			},
		},
		{
			name: "Zero HTTP timeout defaults to 10 seconds",
			envVars: map[string]string{
				"STARFEED_GITHUB_API_TOKEN":   "gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
				"STARFEED_HTTP_TIMEOUT":       "0",
			},
			expectError: false,
			expected: &Config{
				GithubToken:   "gh_token123",
				FreshRssUrl:   "http://freshrss.example.com",
				FreshRssUser:  "testuser",
				FreshRssToken: "freshrss_token456",
				DebugMode:     false,
				SingleRunMode: false,
				HttpTimeout:   10 * time.Second,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockEnvGetter := NewMockEnvGetter(tc.envVars)
			config, err := NewConfig(mockEnvGetter)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if config.GithubToken != tc.expected.GithubToken {
				t.Errorf("Expected GithubToken %s, got %s", tc.expected.GithubToken, config.GithubToken)
			}

			if config.FreshRssUrl != tc.expected.FreshRssUrl {
				t.Errorf("Expected FreshRssUrl %s, got %s", tc.expected.FreshRssUrl, config.FreshRssUrl)
			}

			if config.FreshRssUser != tc.expected.FreshRssUser {
				t.Errorf("Expected FreshRssUser %s, got %s", tc.expected.FreshRssUser, config.FreshRssUser)
			}

			if config.FreshRssToken != tc.expected.FreshRssToken {
				t.Errorf("Expected FreshRssToken %s, got %s", tc.expected.FreshRssToken, config.FreshRssToken)
			}

			if config.DebugMode != tc.expected.DebugMode {
				t.Errorf("Expected DebugMode %t, got %t", tc.expected.DebugMode, config.DebugMode)
			}

			if config.SingleRunMode != tc.expected.SingleRunMode {
				t.Errorf("Expected SingleRunMode %t, got %t", tc.expected.SingleRunMode, config.SingleRunMode)
			}

			if config.HttpTimeout != tc.expected.HttpTimeout {
				t.Errorf("Expected HttpTimeout %v, got %v", tc.expected.HttpTimeout, config.HttpTimeout)
			}
		})
	}
}

func TestOSEnvGetter(t *testing.T) {
	getter := OSEnvGetter{}

	// This should return empty string for a non-existent env var
	result := getter.Getenv("STARFEED_NON_EXISTENT_VAR")
	if result != "" {
		t.Errorf("Expected empty string for non-existent env var, got %s", result)
	}
}

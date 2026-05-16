package config

import (
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/githost"
	"github.com/atomicmeganerd/starfeed/mocks"
)

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
				"STARFEED_GIT_HOST_0":         "github,https://github.com,gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: false,
			expected: &Config{
				GitHosts: []githost.GitHostConfig{
					{Type: githost.GitHub, BaseURL: "https://github.com", Token: "gh_token123"},
				},
				FreshRSSURL:   "http://freshrss.example.com",
				FreshRSSUser:  "testuser",
				FreshRSSToken: "freshrss_token456",
				DebugMode:     false,
				SingleRunMode: false,
				HTTPTimeout:   10 * time.Second,
			},
		},
		{
			name: "All variables present with debug and single run mode",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         "github,https://github.com,gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
				"STARFEED_DEBUG_MODE":         "true",
				"STARFEED_SINGLE_RUN_MODE":    "true",
				"STARFEED_HTTP_TIMEOUT":       "30",
			},
			expectError: false,
			expected: &Config{
				GitHosts: []githost.GitHostConfig{
					{Type: githost.GitHub, BaseURL: "https://github.com", Token: "gh_token123"},
				},
				FreshRSSURL:   "http://freshrss.example.com",
				FreshRSSUser:  "testuser",
				FreshRSSToken: "freshrss_token456",
				DebugMode:     true,
				SingleRunMode: true,
				HTTPTimeout:   30 * time.Second,
			},
		},
		{
			name: "Multiple git hosts",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         "github,https://github.com,gh_token123",
				"STARFEED_GIT_HOST_1":         "forgejo,https://codeberg.org,cb_token456",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: false,
			expected: &Config{
				GitHosts: []githost.GitHostConfig{
					{Type: githost.GitHub, BaseURL: "https://github.com", Token: "gh_token123"},
					{Type: githost.Forgejo, BaseURL: "https://codeberg.org", Token: "cb_token456"},
				},
				FreshRSSURL:   "http://freshrss.example.com",
				FreshRSSUser:  "testuser",
				FreshRSSToken: "freshrss_token456",
				DebugMode:     false,
				SingleRunMode: false,
				HTTPTimeout:   10 * time.Second,
			},
		},
		{
			name: "Spaces in CSV are trimmed",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         " github , https://github.com , gh_token123 ",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: false,
			expected: &Config{
				GitHosts: []githost.GitHostConfig{
					{Type: githost.GitHub, BaseURL: "https://github.com", Token: "gh_token123"},
				},
				FreshRSSURL:   "http://freshrss.example.com",
				FreshRSSUser:  "testuser",
				FreshRSSToken: "freshrss_token456",
				DebugMode:     false,
				SingleRunMode: false,
				HTTPTimeout:   10 * time.Second,
			},
		},
		{
			name: "Missing git hosts should error",
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
				"STARFEED_GIT_HOST_0":         "github,https://github.com,gh_token123",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing FreshRSS user should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         "github,https://github.com,gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing FreshRSS token should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":    "github,https://github.com,gh_token123",
				"STARFEED_FRESHRSS_URL":  "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER": "testuser",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid HTTP timeout defaults to 10 seconds",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         "github,https://github.com,gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
				"STARFEED_HTTP_TIMEOUT":       "invalid",
			},
			expectError: false,
			expected: &Config{
				GitHosts: []githost.GitHostConfig{
					{Type: githost.GitHub, BaseURL: "https://github.com", Token: "gh_token123"},
				},
				FreshRSSURL:   "http://freshrss.example.com",
				FreshRSSUser:  "testuser",
				FreshRSSToken: "freshrss_token456",
				DebugMode:     false,
				SingleRunMode: false,
				HTTPTimeout:   10 * time.Second,
			},
		},
		{
			name: "Zero HTTP timeout defaults to 10 seconds",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         "github,https://github.com,gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
				"STARFEED_HTTP_TIMEOUT":       "0",
			},
			expectError: false,
			expected: &Config{
				GitHosts: []githost.GitHostConfig{
					{Type: githost.GitHub, BaseURL: "https://github.com", Token: "gh_token123"},
				},
				FreshRSSURL:   "http://freshrss.example.com",
				FreshRSSUser:  "testuser",
				FreshRSSToken: "freshrss_token456",
				DebugMode:     false,
				SingleRunMode: false,
				HTTPTimeout:   10 * time.Second,
			},
		},
		{
			name: "Invalid bool for debug mode should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         "github,https://github.com,gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
				"STARFEED_DEBUG_MODE":         "notabool",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid bool for single run mode should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         "github,https://github.com,gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
				"STARFEED_SINGLE_RUN_MODE":    "notabool",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid host type should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         "gitlab,https://gitlab.com,gl_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing CSV parts should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         "github,https://github.com",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty URL in CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         "github,,gh_token123",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty token in CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":         "github,https://github.com,",
				"STARFEED_FRESHRSS_URL":       "http://freshrss.example.com",
				"STARFEED_FRESHRSS_USER":      "testuser",
				"STARFEED_FRESHRSS_API_TOKEN": "freshrss_token456",
			},
			expectError: true,
			expected:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockEnvGetter := mocks.NewMockEnvGetter(tc.envVars)
			cfg, err := NewConfig(mockEnvGetter)

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

			if len(cfg.GitHosts) != len(tc.expected.GitHosts) {
				t.Errorf("Expected %d GitHosts, got %d", len(tc.expected.GitHosts), len(cfg.GitHosts))
				return
			}

			for i, host := range cfg.GitHosts {
				exp := tc.expected.GitHosts[i]
				if host.Type != exp.Type {
					t.Errorf("GitHosts[%d].Type: expected %s, got %s", i, exp.Type, host.Type)
				}
				if host.BaseURL != exp.BaseURL {
					t.Errorf("GitHosts[%d].BaseURL: expected %s, got %s", i, exp.BaseURL, host.BaseURL)
				}
				if host.Token != exp.Token {
					t.Errorf("GitHosts[%d].Token: expected %s, got %s", i, exp.Token, host.Token)
				}
			}

			if cfg.FreshRSSURL != tc.expected.FreshRSSURL {
				t.Errorf("Expected FreshRSSURL %s, got %s", tc.expected.FreshRSSURL, cfg.FreshRSSURL)
			}

			if cfg.FreshRSSUser != tc.expected.FreshRSSUser {
				t.Errorf("Expected FreshRSSUser %s, got %s", tc.expected.FreshRSSUser, cfg.FreshRSSUser)
			}

			if cfg.FreshRSSToken != tc.expected.FreshRSSToken {
				t.Errorf("Expected FreshRSSToken %s, got %s", tc.expected.FreshRSSToken, cfg.FreshRSSToken)
			}

			if cfg.DebugMode != tc.expected.DebugMode {
				t.Errorf("Expected DebugMode %t, got %t", tc.expected.DebugMode, cfg.DebugMode)
			}

			if cfg.SingleRunMode != tc.expected.SingleRunMode {
				t.Errorf("Expected SingleRunMode %t, got %t", tc.expected.SingleRunMode, cfg.SingleRunMode)
			}

			if cfg.HTTPTimeout != tc.expected.HTTPTimeout {
				t.Errorf("Expected HTTPTimeout %v, got %v", tc.expected.HTTPTimeout, cfg.HTTPTimeout)
			}
		})
	}
}

func TestOSEnvGetter(t *testing.T) {
	getter := OSEnvGetter{}

	result := getter.Getenv("STARFEED_NON_EXISTENT_VAR")
	if result != "" {
		t.Errorf("Expected empty string for non-existent env var, got %s", result)
	}
}

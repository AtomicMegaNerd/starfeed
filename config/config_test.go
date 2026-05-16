package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/githost"
	"github.com/atomicmeganerd/starfeed/mocks"
	"github.com/atomicmeganerd/starfeed/rss"
)

var (
	validGitHostCSV = fmt.Sprintf("%s,%s,%s,%s",
		mocks.GitHubType, mocks.GitHubName, mocks.GitHubURL, mocks.GitHubToken)
	validForgejoCSV = fmt.Sprintf("%s,%s,%s,%s",
		mocks.ForgejoType, mocks.ForgejoName, mocks.ForgejoURL, mocks.ForgejoToken)
	validRSSCSV = fmt.Sprintf("%s,%s,%s,%s",
		mocks.FreshRSSType, mocks.FreshRSSURL, mocks.FreshRSSUser, mocks.FreshRSSToken)
	invalidTypeCSV  = fmt.Sprintf("gitlab,mygitlab,https://gitlab.com,%s", mocks.GitHubToken)
	missingPartsCSV = fmt.Sprintf("%s,%s", mocks.GitHubType, mocks.GitHubURL)
	emptyURLCSV     = fmt.Sprintf(
		"%s,%s,,%s", mocks.GitHubType, mocks.GitHubName, mocks.GitHubToken,
	)
	emptyTokenCSV = fmt.Sprintf("%s,%s,%s,", mocks.GitHubType, mocks.GitHubName, mocks.GitHubURL)
	missingRSSCSV = fmt.Sprintf("%s,%s", mocks.FreshRSSType, mocks.FreshRSSURL)

	validGitHostCSVWithSpaces = fmt.Sprintf(" %s , %s , %s , %s ",
		mocks.GitHubType, mocks.GitHubName, mocks.GitHubURL, mocks.GitHubToken)
	validRSSCSVWithSpaces = fmt.Sprintf(" %s , %s , %s , %s ",
		mocks.FreshRSSType, mocks.FreshRSSURL, mocks.FreshRSSUser, mocks.FreshRSSToken)

	emptyURLRSSCSV = fmt.Sprintf("%s,,%s,%s",
		mocks.FreshRSSType, mocks.FreshRSSUser, mocks.FreshRSSToken)
	emptyTokenRSSCSV = fmt.Sprintf("%s,%s,%s,",
		mocks.FreshRSSType, mocks.FreshRSSURL, mocks.FreshRSSUser)

	validGitHostConfig = githost.GitHostConfig{
		Type:    mocks.GitHubType,
		Name:    mocks.GitHubName,
		BaseURL: mocks.GitHubURL,
		Token:   mocks.GitHubToken,
	}

	validForgejoConfig = githost.GitHostConfig{
		Type:    mocks.ForgejoType,
		Name:    mocks.ForgejoName,
		BaseURL: mocks.ForgejoURL,
		Token:   mocks.ForgejoToken,
	}

	validRSSConfig = rss.RSSServerConfig{
		Type:    mocks.FreshRSSType,
		BaseURL: mocks.FreshRSSURL,
		User:    mocks.FreshRSSUser,
		Token:   mocks.FreshRSSToken,
	}
)

func validConfig() *Config {
	return &Config{
		GitHosts:      []githost.GitHostConfig{validGitHostConfig},
		RSSServer:     validRSSConfig,
		DebugMode:     false,
		SingleRunMode: false,
		HTTPTimeout:   10 * time.Second,
	}
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
				"STARFEED_GIT_HOST_0": validGitHostCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: false,
			expected:    validConfig(),
		},
		{
			name: "All variables present with debug and single run mode",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":      validGitHostCSV,
				"STARFEED_RSS_SERVER":      validRSSCSV,
				"STARFEED_DEBUG_MODE":      mocks.TrueBool,
				"STARFEED_SINGLE_RUN_MODE": mocks.TrueBool,
				"STARFEED_HTTP_TIMEOUT":    mocks.ValidTimeout,
			},
			expectError: false,
			expected: &Config{
				GitHosts:      []githost.GitHostConfig{validGitHostConfig},
				RSSServer:     validRSSConfig,
				DebugMode:     true,
				SingleRunMode: true,
				HTTPTimeout:   30 * time.Second,
			},
		},
		{
			name: "Multiple git hosts",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitHostCSV,
				"STARFEED_GIT_HOST_1": validForgejoCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: false,
			expected: &Config{
				GitHosts:      []githost.GitHostConfig{validGitHostConfig, validForgejoConfig},
				RSSServer:     validRSSConfig,
				DebugMode:     false,
				SingleRunMode: false,
				HTTPTimeout:   10 * time.Second,
			},
		},
		{
			name: "Spaces in CSV are trimmed",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitHostCSVWithSpaces,
				"STARFEED_RSS_SERVER": validRSSCSVWithSpaces,
			},
			expectError: false,
			expected:    validConfig(),
		},
		{
			name: "Missing git hosts should error",
			envVars: map[string]string{
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing RSS server should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitHostCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid HTTP timeout defaults to 10 seconds",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":   validGitHostCSV,
				"STARFEED_RSS_SERVER":   validRSSCSV,
				"STARFEED_HTTP_TIMEOUT": mocks.Invalid,
			},
			expectError: false,
			expected:    validConfig(),
		},
		{
			name: "Zero HTTP timeout defaults to 10 seconds",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":   validGitHostCSV,
				"STARFEED_RSS_SERVER":   validRSSCSV,
				"STARFEED_HTTP_TIMEOUT": mocks.ZeroTimeout,
			},
			expectError: false,
			expected:    validConfig(),
		},
		{
			name: "Invalid bool for debug mode should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitHostCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
				"STARFEED_DEBUG_MODE": mocks.Invalid,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid bool for single run mode should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":      validGitHostCSV,
				"STARFEED_RSS_SERVER":      validRSSCSV,
				"STARFEED_SINGLE_RUN_MODE": mocks.Invalid,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid host type should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": invalidTypeCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing CSV parts for git host should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": missingPartsCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty URL in git host CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": emptyURLCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty token in git host CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": emptyTokenCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing CSV parts for RSS server should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitHostCSV,
				"STARFEED_RSS_SERVER": missingRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty URL in RSS server CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitHostCSV,
				"STARFEED_RSS_SERVER": emptyURLRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty token in RSS server CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitHostCSV,
				"STARFEED_RSS_SERVER": emptyTokenRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockEnvGetter := NewMockEnvGetter(tc.envVars)
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
				t.Errorf(
					"Expected %d GitHosts, got %d",
					len(tc.expected.GitHosts),
					len(cfg.GitHosts),
				)
				return
			}

			for i, host := range cfg.GitHosts {
				exp := tc.expected.GitHosts[i]
				if host.Type != exp.Type {
					t.Errorf("GitHosts[%d].Type: expected %s, got %s", i, exp.Type, host.Type)
				}
				if host.Name != exp.Name {
					t.Errorf("GitHosts[%d].Name: expected %s, got %s", i, exp.Name, host.Name)
				}
				if host.BaseURL != exp.BaseURL {
					t.Errorf(
						"GitHosts[%d].BaseURL: expected %s, got %s", i, exp.BaseURL, host.BaseURL,
					)
				}
				if host.Token != exp.Token {
					t.Errorf("GitHosts[%d].Token: expected %s, got %s", i, exp.Token, host.Token)
				}
			}

			if cfg.RSSServer.Type != tc.expected.RSSServer.Type {
				t.Errorf(
					"Expected RSSServer.Type %s, got %s",
					tc.expected.RSSServer.Type, cfg.RSSServer.Type,
				)
			}
			if cfg.RSSServer.BaseURL != tc.expected.RSSServer.BaseURL {
				t.Errorf(
					"Expected RSSServer.URL %s, got %s",
					tc.expected.RSSServer.BaseURL, cfg.RSSServer.BaseURL,
				)
			}
			if cfg.RSSServer.User != tc.expected.RSSServer.User {
				t.Errorf(
					"Expected RSSServer.User %s, got %s",
					tc.expected.RSSServer.User, cfg.RSSServer.User,
				)
			}
			if cfg.RSSServer.Token != tc.expected.RSSServer.Token {
				t.Errorf("Expected RSSServer.Token %s, got %s",
					tc.expected.RSSServer.Token, cfg.RSSServer.Token)
			}

			if cfg.DebugMode != tc.expected.DebugMode {
				t.Errorf("Expected DebugMode %t, got %t", tc.expected.DebugMode, cfg.DebugMode)
			}

			if cfg.SingleRunMode != tc.expected.SingleRunMode {
				t.Errorf(
					"Expected SingleRunMode %t, got %t",
					tc.expected.SingleRunMode, cfg.SingleRunMode,
				)
			}

			if cfg.HTTPTimeout != tc.expected.HTTPTimeout {
				t.Errorf(
					"Expected HTTPTimeout %v, got %v", tc.expected.HTTPTimeout, cfg.HTTPTimeout,
				)
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

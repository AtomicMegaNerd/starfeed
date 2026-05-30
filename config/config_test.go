package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/mocks"
)

var (
	validGitHostCSV = fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s",
		GitHubHostType,
		mocks.GitHubName,
		mocks.GitHubURL,
		mocks.GitHubAPIURL,
		mocks.GitHubToken,
		mocks.TrueBool,
	)
	validForgejoCSV = fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s",
		ForgejoHostType,
		mocks.ForgejoName,
		mocks.ForgejoURL,
		mocks.ForgejoAPIURL,
		mocks.ForgejoToken,
		mocks.TrueBool,
	)
	validRSSCSV = fmt.Sprintf(
		"%s,%s,%s,%s,%s",
		mocks.FreshRSSType,
		mocks.FreshRSSURL,
		mocks.FreshRSSUser,
		mocks.FreshRSSToken,
		mocks.TrueBool,
	)
	invalidTypeCSV = fmt.Sprintf(
		"gitlab,mygitlab,https://gitlab.com,https://api.gitlab.com,%s,%s",
		mocks.GitHubToken, mocks.TrueBool,
	)
	missingPartsCSV = fmt.Sprintf("%s,%s", GitHubHostType, mocks.GitHubURL)
	emptyURLCSV     = fmt.Sprintf(
		"%s,%s,,%s,%s,%s",
		GitHubHostType,
		mocks.GitHubName,
		mocks.GitHubAPIURL,
		mocks.GitHubToken,
		mocks.TrueBool,
	)

	emptyTokenCSV = fmt.Sprintf(
		"%s,%s,%s,%s,,%s",
		GitHubHostType,
		mocks.GitHubName,
		mocks.GitHubURL,
		mocks.GitHubAPIURL,
		mocks.TrueBool,
	)
	missingRSSCSV = fmt.Sprintf("%s,%s", mocks.FreshRSSType, mocks.FreshRSSURL)

	emptyBaseURLCSV = fmt.Sprintf(
		"%s,%s,,%s,%s,%s",
		GitHubHostType,
		mocks.GitHubName,
		mocks.GitHubAPIURL,
		mocks.GitHubToken,
		mocks.TrueBool,
	)
	emptyAPIURLCSV = fmt.Sprintf(
		"%s,%s,%s,,%s,%s",
		GitHubHostType,
		mocks.GitHubName,
		mocks.GitHubURL,
		mocks.GitHubToken,
		mocks.TrueBool,
	)

	validGitHostCSVWithSpaces = fmt.Sprintf(
		" %s , %s , %s , %s , %s , %s ",
		GitHubHostType,
		mocks.GitHubName,
		mocks.GitHubURL,
		mocks.GitHubAPIURL,
		mocks.GitHubToken,
		mocks.TrueBool,
	)
	validRSSCSVWithSpaces = fmt.Sprintf(
		" %s , %s , %s , %s , %s ",
		mocks.FreshRSSType,
		mocks.FreshRSSURL,
		mocks.FreshRSSUser,
		mocks.FreshRSSToken,
		mocks.TrueBool,
	)

	emptyURLRSSCSV = fmt.Sprintf("%s,,%s,%s,%s",
		mocks.FreshRSSType, mocks.FreshRSSUser, mocks.FreshRSSToken, mocks.TrueBool)
	emptyTokenRSSCSV = fmt.Sprintf("%s,%s,%s,,%s",
		mocks.FreshRSSType, mocks.FreshRSSURL, mocks.FreshRSSUser, mocks.TrueBool)
	invalidEnabledGitHostCSV = fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s",
		GitHubHostType,
		mocks.GitHubName,
		mocks.GitHubURL,
		mocks.GitHubAPIURL,
		mocks.GitHubToken,
		mocks.Invalid,
	)
	invalidEnabledRSSCSV = fmt.Sprintf(
		"%s,%s,%s,%s,%s",
		mocks.FreshRSSType,
		mocks.FreshRSSURL,
		mocks.FreshRSSUser,
		mocks.FreshRSSToken,
		mocks.Invalid,
	)
	missingEnabledGitHostCSV = fmt.Sprintf("%s,%s,%s,%s,%s",
		GitHubHostType, mocks.GitHubName, mocks.GitHubURL, mocks.GitHubAPIURL, mocks.GitHubToken)
	missingEnabledRSSCSV = fmt.Sprintf("%s,%s,%s,%s",
		mocks.FreshRSSType, mocks.FreshRSSURL, mocks.FreshRSSUser, mocks.FreshRSSToken)
)

func validConfig() *Config {
	return &Config{
		GitHostConfigs:  []GitHostConfig{MockValidGitHubConfig},
		RSSServerConfig: MockValidFreshRSSConfig,
		DebugMode:       false,
		SingleRunMode:   false,
		HTTPTimeout:     defaultHttpTimeoutSeconds * time.Second,
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
				GitHostConfigs:  []GitHostConfig{MockValidGitHubConfig},
				RSSServerConfig: MockValidFreshRSSConfig,
				DebugMode:       true,
				SingleRunMode:   true,
				HTTPTimeout:     30 * time.Second,
			},
		},
		{
			name: "Multiple Git hosts",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitHostCSV,
				"STARFEED_GIT_HOST_1": validForgejoCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: false,
			expected: &Config{
				GitHostConfigs:  []GitHostConfig{MockValidGitHubConfig, MockValidForgejoConfig},
				RSSServerConfig: MockValidFreshRSSConfig,
				DebugMode:       false,
				SingleRunMode:   false,
				HTTPTimeout:     defaultHttpTimeoutSeconds * time.Second,
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
			name: "Missing Git hosts should error",
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
		{
			name: "Empty BaseURL in git host CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": emptyBaseURLCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty ApiURL in git host CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": emptyAPIURLCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid Enabled in git host CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": invalidEnabledGitHostCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid Enabled in RSS server CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitHostCSV,
				"STARFEED_RSS_SERVER": invalidEnabledRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing Enabled in git host CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": missingEnabledGitHostCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing Enabled in RSS server CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitHostCSV,
				"STARFEED_RSS_SERVER": missingEnabledRSSCSV,
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

			if len(cfg.GitHostConfigs) != len(tc.expected.GitHostConfigs) {
				t.Errorf(
					"Expected %d GitHosts, got %d",
					len(tc.expected.GitHostConfigs),
					len(cfg.GitHostConfigs),
				)
				return
			}

			// NOTE: The validators in githost.GitHosts and rss.RSSServer constructor functions
			// will validate that all of the values we pass in are valid. We'll test those
			// factory functions in their respective packages.

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

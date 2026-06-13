package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/testutils"
)

var (
	validGitForgetCSV = fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s",
		GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubURL,
		testutils.GitHubAPIURL,
		testutils.GitHubToken,
		testutils.TrueBool,
	)
	validForgejoCSV = fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s",
		ForgejoForgeType,
		testutils.CodebergName,
		testutils.CodebergURL,
		testutils.CodebergAPIURL,
		testutils.CodebergToken,
		testutils.TrueBool,
	)
	validRSSCSV = fmt.Sprintf(
		"%s,%s,%s,%s,%s",
		FreshRSSType,
		testutils.FreshRSSURL,
		testutils.FreshRSSUser,
		testutils.FreshRSSToken,
		testutils.TrueBool,
	)
	invalidTypeCSV = fmt.Sprintf(
		"gitlab,mygitlab,https://gitlab.com,https://api.gitlab.com,%s,%s",
		testutils.GitHubToken, testutils.TrueBool,
	)
	missingPartsCSV = fmt.Sprintf("%s,%s", GitHubForgeType, testutils.GitHubURL)
	emptyURLCSV     = fmt.Sprintf(
		"%s,%s,,%s,%s,%s",
		GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubAPIURL,
		testutils.GitHubToken,
		testutils.TrueBool,
	)

	emptyTokenCSV = fmt.Sprintf(
		"%s,%s,%s,%s,,%s",
		GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubURL,
		testutils.GitHubAPIURL,
		testutils.TrueBool,
	)
	missingRSSCSV = fmt.Sprintf("%s,%s", FreshRSSType, testutils.FreshRSSURL)

	emptyBaseURLCSV = fmt.Sprintf(
		"%s,%s,,%s,%s,%s",
		GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubAPIURL,
		testutils.GitHubToken,
		testutils.TrueBool,
	)
	emptyAPIURLCSV = fmt.Sprintf(
		"%s,%s,%s,,%s,%s",
		GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubURL,
		testutils.GitHubToken,
		testutils.TrueBool,
	)

	validGitForgeCSVWithSpaces = fmt.Sprintf(
		" %s , %s , %s , %s , %s , %s ",
		GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubURL,
		testutils.GitHubAPIURL,
		testutils.GitHubToken,
		testutils.TrueBool,
	)
	validRSSCSVWithSpaces = fmt.Sprintf(
		" %s , %s , %s , %s , %s ",
		FreshRSSType,
		testutils.FreshRSSURL,
		testutils.FreshRSSUser,
		testutils.FreshRSSToken,
		testutils.TrueBool,
	)

	emptyURLRSSCSV = fmt.Sprintf("%s,,%s,%s,%s",
		FreshRSSType, testutils.FreshRSSUser, testutils.FreshRSSToken, testutils.TrueBool)
	emptyTokenRSSCSV = fmt.Sprintf("%s,%s,%s,,%s",
		FreshRSSType, testutils.FreshRSSURL, testutils.FreshRSSUser, testutils.TrueBool)
	invalidEnabledGitForgeCSV = fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s",
		GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubURL,
		testutils.GitHubAPIURL,
		testutils.GitHubToken,
		testutils.Invalid,
	)
	invalidEnabledRSSCSV = fmt.Sprintf(
		"%s,%s,%s,%s,%s",
		FreshRSSType,
		testutils.FreshRSSURL,
		testutils.FreshRSSUser,
		testutils.FreshRSSToken,
		testutils.Invalid,
	)
	missingEnabledGitForgeCSV = fmt.Sprintf(
		"%s,%s,%s,%s,%s",
		GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubURL,
		testutils.GitHubAPIURL,
		testutils.GitHubToken,
	)
	missingEnabledRSSCSV = fmt.Sprintf("%s,%s,%s,%s",
		FreshRSSType, testutils.FreshRSSURL, testutils.FreshRSSUser, testutils.FreshRSSToken)
)

func validConfig() *Config {
	return &Config{
		GitForgeConfigs: []GitForgeConfig{MockValidGitHubConfig},
		RSSServerConfig: MockValidFreshRSSConfig,
		DebugMode:       false,
		SingleRunMode:   false,
		HTTPTimeout:     defaultHTTPTimeoutSeconds * time.Second,
	}
}

func TestNewConfig(t *testing.T) {
	testCases := []struct {
		name        string
		envVars     map[string]string
		expectError bool
		expected    *Config
	}{
		{
			name: "All required variables present",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitForgetCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: false,
			expected:    validConfig(),
		},
		{
			name: "All variables present with debug and single run mode",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":      validGitForgetCSV,
				"STARFEED_RSS_SERVER":      validRSSCSV,
				"STARFEED_DEBUG_MODE":      testutils.TrueBool,
				"STARFEED_SINGLE_RUN_MODE": testutils.TrueBool,
				"STARFEED_HTTP_TIMEOUT":    testutils.ValidTimeout,
			},
			expectError: false,
			expected: &Config{
				GitForgeConfigs: []GitForgeConfig{MockValidGitHubConfig},
				RSSServerConfig: MockValidFreshRSSConfig,
				DebugMode:       true,
				SingleRunMode:   true,
				HTTPTimeout:     30 * time.Second,
			},
		},
		{
			name: "Multiple Git hosts",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitForgetCSV,
				"STARFEED_GIT_HOST_1": validForgejoCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: false,
			expected: &Config{
				GitForgeConfigs: []GitForgeConfig{MockValidGitHubConfig, MockValidCodebergConfig},
				RSSServerConfig: MockValidFreshRSSConfig,
				DebugMode:       false,
				SingleRunMode:   false,
				HTTPTimeout:     defaultHTTPTimeoutSeconds * time.Second,
			},
		},
		{
			name: "Spaces in CSV are trimmed",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitForgeCSVWithSpaces,
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
				"STARFEED_GIT_HOST_0": validGitForgetCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid HTTP timeout defaults to 10 seconds",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":   validGitForgetCSV,
				"STARFEED_RSS_SERVER":   validRSSCSV,
				"STARFEED_HTTP_TIMEOUT": testutils.Invalid,
			},
			expectError: false,
			expected:    validConfig(),
		},
		{
			name: "Zero HTTP timeout defaults to 10 seconds",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":   validGitForgetCSV,
				"STARFEED_RSS_SERVER":   validRSSCSV,
				"STARFEED_HTTP_TIMEOUT": testutils.ZeroTimeout,
			},
			expectError: false,
			expected:    validConfig(),
		},
		{
			name: "Invalid bool for debug mode should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitForgetCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
				"STARFEED_DEBUG_MODE": testutils.Invalid,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid bool for single run mode should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0":      validGitForgetCSV,
				"STARFEED_RSS_SERVER":      validRSSCSV,
				"STARFEED_SINGLE_RUN_MODE": testutils.Invalid,
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
				"STARFEED_GIT_HOST_0": validGitForgetCSV,
				"STARFEED_RSS_SERVER": missingRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty URL in RSS server CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitForgetCSV,
				"STARFEED_RSS_SERVER": emptyURLRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty token in RSS server CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitForgetCSV,
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
				"STARFEED_GIT_HOST_0": invalidEnabledGitForgeCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid Enabled in RSS server CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitForgetCSV,
				"STARFEED_RSS_SERVER": invalidEnabledRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing Enabled in git host CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": missingEnabledGitForgeCSV,
				"STARFEED_RSS_SERVER": validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing Enabled in RSS server CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_HOST_0": validGitForgetCSV,
				"STARFEED_RSS_SERVER": missingEnabledRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			mockEnvGetter := testutils.NewMockEnvGetter(tc.envVars)
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

			if len(cfg.GitForgeConfigs) != len(tc.expected.GitForgeConfigs) {
				t.Errorf(
					"Expected %d GitForges, got %d",
					len(tc.expected.GitForgeConfigs),
					len(cfg.GitForgeConfigs),
				)
				return
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

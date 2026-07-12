package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/atomicmeganerd/starfeed/testutils"
)

var (
	MockRSSConfig = rss.RSSServerConfig{
		Name: testutils.FreshRSSName,
		URL:  testutils.FreshRSSURL,
		User: testutils.FreshRSSUser,
	}
	MockGitHubConfig = gitforge.GitForgeConfig{
		Type:  gitforge.GitHubForgeType,
		Name:  testutils.GitHubName,
		Fqdn:  testutils.GitHubFqdn,
		Token: testutils.GitHubToken,
	}
	MockCodebergConfig = gitforge.GitForgeConfig{
		Type:  gitforge.ForgejoForgeType,
		Name:  testutils.CodebergName,
		Fqdn:  testutils.CodebergFqdn,
		Token: testutils.CodebergToken,
	}

	validGitHubCSV = fmt.Sprintf(
		"%s,%s,%s,%s",
		gitforge.GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubFqdn,
		testutils.GitHubToken,
	)
	validForgejoCSV = fmt.Sprintf(
		"%s,%s,%s,%s",
		gitforge.ForgejoForgeType,
		testutils.CodebergName,
		testutils.CodebergFqdn,
		testutils.CodebergToken,
	)
	validRSSCSV = fmt.Sprintf(
		"%s,%s,%s,%s",
		testutils.FreshRSSName,
		testutils.FreshRSSURL,
		testutils.FreshRSSUser,
		testutils.FreshRSSToken,
	)
	missingPartsCSV = fmt.Sprintf("%s,%s", gitforge.GitHubForgeType, testutils.GitHubFqdn)
	emptyURLCSV     = fmt.Sprintf(
		"%s,%s,,%s",
		gitforge.GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubToken,
	)

	emptyTokenCSV = fmt.Sprintf(
		"%s,%s,%s,",
		gitforge.GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubFqdn,
	)
	missingRSSCSV = fmt.Sprintf("%s,%s", testutils.FreshRSSName, testutils.FreshRSSURL)

	emptyAPIURLCSV = fmt.Sprintf(
		"%s,%s,,%s",
		gitforge.GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubToken,
	)

	validGitForgeCSVWithSpaces = fmt.Sprintf(
		" %s , %s , %s , %s",
		gitforge.GitHubForgeType,
		testutils.GitHubName,
		testutils.GitHubFqdn,
		testutils.GitHubToken,
	)
	validRSSCSVWithSpaces = fmt.Sprintf(
		" %s , %s , %s , %s",
		testutils.FreshRSSName,
		testutils.FreshRSSURL,
		testutils.FreshRSSUser,
		testutils.FreshRSSToken,
	)

	emptyURLRSSCSV = fmt.Sprintf("%s,,%s,%s",
		testutils.FreshRSSName, testutils.FreshRSSUser, testutils.FreshRSSToken)
	emptyTokenRSSCSV = fmt.Sprintf("%s,%s,%s,",
		testutils.FreshRSSName, testutils.FreshRSSURL, testutils.FreshRSSUser)
)

func validConfig() *Config {
	return &Config{
		GitForgeConfigs: []gitforge.GitForgeConfig{MockGitHubConfig},
		RSSServerConfig: MockRSSConfig,
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
				"STARFEED_GIT_FORGE_0": validGitHubCSV,
				"STARFEED_RSS_SERVER":  validRSSCSV,
			},
			expectError: false,
			expected:    validConfig(),
		},
		{
			name: "All variables present with debug and single run mode",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0":     validGitHubCSV,
				"STARFEED_RSS_SERVER":      validRSSCSV,
				"STARFEED_DEBUG_MODE":      testutils.TrueBool,
				"STARFEED_SINGLE_RUN_MODE": testutils.TrueBool,
				"STARFEED_HTTP_TIMEOUT":    testutils.ValidTimeout,
			},
			expectError: false,
			expected: &Config{
				GitForgeConfigs: []gitforge.GitForgeConfig{MockGitHubConfig},
				RSSServerConfig: MockRSSConfig,
				DebugMode:       true,
				SingleRunMode:   true,
				HTTPTimeout:     30 * time.Second,
			},
		},
		{
			name: "Multiple Git hosts",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0": validGitHubCSV,
				"STARFEED_GIT_FORGE_1": validForgejoCSV,
				"STARFEED_RSS_SERVER":  validRSSCSV,
			},
			expectError: false,
			expected: &Config{
				GitForgeConfigs: []gitforge.GitForgeConfig{
					MockGitHubConfig,
					MockCodebergConfig,
				},
				RSSServerConfig: MockRSSConfig,
				DebugMode:       false,
				SingleRunMode:   false,
				HTTPTimeout:     defaultHTTPTimeoutSeconds * time.Second,
			},
		},
		{
			name: "Spaces in CSV are trimmed",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0": validGitForgeCSVWithSpaces,
				"STARFEED_RSS_SERVER":  validRSSCSVWithSpaces,
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
				"STARFEED_GIT_FORGE_0": validGitHubCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid HTTP timeout defaults to 10 seconds",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0":  validGitHubCSV,
				"STARFEED_RSS_SERVER":   validRSSCSV,
				"STARFEED_HTTP_TIMEOUT": testutils.Invalid,
			},
			expectError: false,
			expected:    validConfig(),
		},
		{
			name: "Zero HTTP timeout defaults to 10 seconds",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0":  validGitHubCSV,
				"STARFEED_RSS_SERVER":   validRSSCSV,
				"STARFEED_HTTP_TIMEOUT": testutils.ZeroTimeout,
			},
			expectError: false,
			expected:    validConfig(),
		},
		{
			name: "Invalid bool for debug mode should error",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0": validGitHubCSV,
				"STARFEED_RSS_SERVER":  validRSSCSV,
				"STARFEED_DEBUG_MODE":  testutils.Invalid,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Invalid bool for single run mode should error",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0":     validGitHubCSV,
				"STARFEED_RSS_SERVER":      validRSSCSV,
				"STARFEED_SINGLE_RUN_MODE": testutils.Invalid,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing CSV parts for git host should error",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0": missingPartsCSV,
				"STARFEED_RSS_SERVER":  validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty URL in git host CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0": emptyURLCSV,
				"STARFEED_RSS_SERVER":  validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty token in git host CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0": emptyTokenCSV,
				"STARFEED_RSS_SERVER":  validRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Missing CSV parts for RSS server should error",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0": validGitHubCSV,
				"STARFEED_RSS_SERVER":  missingRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty URL in RSS server CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0": validGitHubCSV,
				"STARFEED_RSS_SERVER":  emptyURLRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty token in RSS server CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0": validGitHubCSV,
				"STARFEED_RSS_SERVER":  emptyTokenRSSCSV,
			},
			expectError: true,
			expected:    nil,
		},
		{
			name: "Empty ApiURL in git host CSV should error",
			envVars: map[string]string{
				"STARFEED_GIT_FORGE_0": emptyAPIURLCSV,
				"STARFEED_RSS_SERVER":  validRSSCSV,
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

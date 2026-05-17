package config

import (
	"github.com/atomicmeganerd/starfeed/mocks"
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

var (
	MockValidGitHubConfig = GitHostConfig{
		mocks.GitHubType,
		mocks.GitHubName,
		mocks.GitHubURL,
		mocks.GitHubAPIURL,
		mocks.GitHubToken,
		true,
	}

	MockValidForgejoConfig = GitHostConfig{
		mocks.ForgejoType,
		mocks.ForgejoName,
		mocks.ForgejoURL,
		mocks.ForgejoAPIUrl,
		mocks.ForgejoToken,
		true,
	}

	MockValidFreshRSSConfig = &RSSServerConfig{
		mocks.FreshRSSType, mocks.FreshRSSURL, mocks.FreshRSSUser, mocks.FreshRSSToken, true,
	}
)

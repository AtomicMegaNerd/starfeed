package config

import "github.com/atomicmeganerd/starfeed/testutils"

var (
	MockValidCodebergConfig = GitForgeConfig{
		Type:    ForgejoForgeType,
		Name:    testutils.CodebergName,
		BaseURL: testutils.CodebergURL,
		ApiURL:  testutils.CodebergAPIURL,
		Token:   testutils.GitHubToken,
	}

	MockValidGitHubConfig = GitForgeConfig{
		Type:    GitHubForgeType,
		Name:    testutils.GitHubName,
		BaseURL: testutils.GitHubURL,
		ApiURL:  testutils.GitHubAPIURL,
		Token:   testutils.GitHubToken,
	}

	MockValidFreshRSSConfig = RSSServerConfig{
		Type:    FreshRSSType,
		BaseURL: testutils.FreshRSSURL,
		User:    testutils.FreshRSSUser,
		Enabled: false,
	}

	MockValidFreshRSSEnabledConfig = RSSServerConfig{
		Type:    FreshRSSType,
		BaseURL: testutils.FreshRSSURL,
		User:    testutils.FreshRSSUser,
		Enabled: true,
	}
)

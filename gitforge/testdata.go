package gitforge

import (
	"log/slog"
	"net/http"

	"github.com/atomicmeganerd/starfeed/testutils"
)

func MockValidGitHub(client *http.Client, logger *slog.Logger) GitForge {
	return NewGitForge(MockValidGitHubConfig, logger, client)
}

func MockValidCodeberg(client *http.Client, logger *slog.Logger) GitForge {
	return NewGitForge(MockValidCodebergConfig, logger, client)
}

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
)

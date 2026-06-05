package githost

import (
	"log/slog"
	"net/http"
	"regexp"

	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/mocks"
)

var (
	GithubRelRegex = regexp.MustCompile(
		`^https://github\.com/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
	)
	CodebergRelRegex = regexp.MustCompile(
		`^https://codeberg\.org/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
	)

	MockValidGitHub = func(client *http.Client, logger *slog.Logger) GitHost {
		return GitHost{
			Name:                 mocks.GitHubName,
			Enabled:              true,
			HostType:             config.GitHubHostType,
			ReleaseFeedPattern:   GithubRelRegex,
			logger:               logger,
			client:               client,
			starredReposFetchURL: "https://api.github.com",
		}
	}

	MockValidCodeberg = func(client *http.Client, logger *slog.Logger) GitHost {
		return GitHost{
			Name:                 mocks.ForgejoName,
			Enabled:              true,
			HostType:             config.ForgejoHostType,
			ReleaseFeedPattern:   CodebergRelRegex,
			logger:               logger,
			client:               client,
			starredReposFetchURL: "https://api.forgejo.org",
		}
	}
)

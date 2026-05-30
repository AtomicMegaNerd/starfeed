package githost

import (
	"io"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/mocks"
)

var (
	nextPageLinkRegex = regexp.MustCompile(`<([^>]+)>; rel="next"`)
	githubRelRegex    = regexp.MustCompile(
		`^https://github\.com/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
	)
	forgejoRelRegex = regexp.MustCompile(
		`^https://codeberg\.org/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
	)

	MockValidGitHub = func(client *http.Client) GitHost {
		return GitHost{
			hostType:         config.GitHubHostType,
			Name:             mocks.GitHubName,
			baseURL:          mocks.GitHubURL,
			token:            mocks.GitHubToken,
			Enabled:          true,
			logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
			client:           client,
			getReposURL:      "https://api.github.com",
			nextPagePattern:  nextPageLinkRegex,
			isReleasePattern: githubRelRegex,
			headers: map[string]string{
				"Authorization":        "Bearer " + mocks.GitHubToken,
				"X-GitHub-Api-Version": "2022-11-28",
				"User-Agent":           "github.com/atomicmeganerd/starfeed",
				"Content-Type":         "application/json",
				"Accept":               "application/json",
			},
		}
	}

	MockValidForgejo = func(client *http.Client) GitHost {
		return GitHost{
			hostType:         config.ForgejoHostType,
			Name:             mocks.ForgejoName,
			baseURL:          mocks.ForgejoURL,
			token:            mocks.ForgejoToken,
			Enabled:          true,
			logger:           slog.New(slog.NewTextHandler(io.Discard, nil)),
			client:           client,
			getReposURL:      "https://api.forgejo.org",
			nextPagePattern:  nextPageLinkRegex,
			isReleasePattern: forgejoRelRegex,
			headers: map[string]string{
				"Authorization": "Bearer " + mocks.ForgejoToken,
				"User-Agent":    "github.com/atomicmeganerd/starfeed",
				"Content-Type":  "application/json",
				"Accept":        "application/json",
			},
		}
	}
)

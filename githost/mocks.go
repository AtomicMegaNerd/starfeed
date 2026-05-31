package githost

import (
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
			logger:           mocks.TestLogger(),
			client:           client,
			getReposURL:      "https://api.github.com",
			nextPagePattern:  nextPageLinkRegex,
			isReleasePattern: githubRelRegex,
		}
	}

	MockValidForgejo = func(client *http.Client) GitHost {
		return GitHost{
			hostType:         config.ForgejoHostType,
			Name:             mocks.ForgejoName,
			baseURL:          mocks.ForgejoURL,
			token:            mocks.ForgejoToken,
			Enabled:          true,
			logger:           mocks.TestLogger(),
			client:           client,
			getReposURL:      "https://api.forgejo.org",
			nextPagePattern:  nextPageLinkRegex,
			isReleasePattern: forgejoRelRegex,
		}
	}
)

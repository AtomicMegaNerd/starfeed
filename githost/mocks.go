package githost

import (
	"net/http"
	"regexp"

	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/mocks"
)

var (
	githubRelRegex = regexp.MustCompile(
		`^https://github\.com/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
	)
	forgejoRelRegex = regexp.MustCompile(
		`^https://codeberg\.org/[\w\.\-]+/[\w\.\-]+/releases\.atom`,
	)

	MockValidGitHub = func(client *http.Client) GitHost {
		return GitHost{
			hostType:         config.GitHubHostType,
			Name:             mocks.GitHubName,
			Enabled:          true,
			logger:           mocks.TestLogger(),
			client:           client,
			getReposURL:      "https://api.github.com",
			isReleasePattern: githubRelRegex,
		}
	}

	MockValidForgejo = func(client *http.Client) GitHost {
		return GitHost{
			hostType:         config.ForgejoHostType,
			Name:             mocks.ForgejoName,
			Enabled:          true,
			logger:           mocks.TestLogger(),
			client:           client,
			getReposURL:      "https://api.forgejo.org",
			isReleasePattern: forgejoRelRegex,
		}
	}
)

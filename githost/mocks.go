package githost

import (
	"net/http"
	"regexp"

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
		return &gitHost{
			hostType:         mocks.GitHubType,
			name:             mocks.GitHubName,
			baseURL:          mocks.GitHubURL,
			token:            mocks.GitHubToken,
			client:           client,
			GetReposURL:      "",
			NextPagePattern:  nextPageLinkRegex,
			IsReleasePattern: githubRelRegex,
			Headers: map[string]string{
				"Authorization":        "Bearer " + mocks.GitHubToken,
				"X-GitHub-Api-Version": "2022-11-28",
				"User-Agent":           "github.com/atomicmeganerd/starfeed",
				"Content-Type":         "application/json",
				"Accept":               "application/json",
			},
		}
	}

	MockValidForgejo = func(client *http.Client) GitHost {
		return &gitHost{
			hostType:         mocks.ForgejoType,
			name:             mocks.ForgejoName,
			baseURL:          mocks.ForgejoURL,
			token:            mocks.ForgejoToken,
			client:           client,
			GetReposURL:      "",
			NextPagePattern:  nextPageLinkRegex,
			IsReleasePattern: forgejoRelRegex,
			Headers: map[string]string{
				"Authorization": "Bearer " + mocks.ForgejoToken,
				"User-Agent":    "github.com/atomicmeganerd/starfeed",
				"Content-Type":  "application/json",
				"Accept":        "application/json",
			},
		}
	}
)

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/atomicmeganerd/starfeed/githost"
)

// gitHub is the private implementation of GitHubStarredFeedBuilder.
type gitHub struct {
	gitHost           githost.GitHostConfig
	client            *http.Client
	headers           map[string]string
	nextPageLinkRegex *regexp.Regexp
	isRelRepoRegex    *regexp.Regexp
}

// NewGitHub creates a new GitHubStarredFeedBuilder instance.
// Arguments:
// - cfg: This holds the configuration state this object needs.
// - client: The http client to use for requests (used for mocking).
func NewGitHub(
	gitHost githost.GitHostConfig,
	client *http.Client,
) githost.GitHost {
	// This regex is used to find the next page link in the GitHub API response
	nextPageLinkRegex := regexp.MustCompile(`<([^>]+)>; rel="next"`)
	// This regex is used to determine if an RSS feed is a GitHub release feed
	isRelRepoRegex := regexp.MustCompile(`^https://github.com/[\w\.\-]+/[\w\.\-]+/releases\.atom`)

	headers := map[string]string{
		"Authorization":        fmt.Sprintf("Bearer %s", gitHost.Token),
		"X-GitHub-Api-Version": "2022-11-28",
		"User-Agent":           "github.com/atomicmeganerd/starfeed",
		"Content-Type":         "application/json",
		"Accept":               "application/json",
	}
	return &gitHub{gitHost, client, headers, nextPageLinkRegex, isRelRepoRegex}
}

// This will return all starred repos including the Atom feeds for their releases
// It returns a map of releaseFeedUrl -> GitHubRepo
func (gh *gitHub) GetStarredRepos(
	ctx context.Context,
) (map[string]githost.Repo, error) {
	allFeeds := make(map[string]githost.Repo)
	getUrl := "https://api.github.com/user/starred?per_page=100"
	slog.Debug("Querying GitHub for starred repos", "url", getUrl)

	for {
		ghResponse, err := githost.DoApiRequest(
			ctx,
			getUrl,
			gh.headers,
			gh.nextPageLinkRegex,
			gh.client,
		)
		if err != nil {
			return nil, err
		}

		var repos []githost.BaseRepo
		if err = json.Unmarshal(ghResponse.Data, &repos); err != nil {
			return nil, err
		}

		for _, repo := range repos {
			allFeeds[repo.FeedURL()] = &repo
		}

		// If there is no next page we are done...
		if ghResponse.NextPage == "" {
			return allFeeds, nil
		}

		slog.Debug("Found next page", "url", ghResponse.NextPage)
		getUrl = ghResponse.NextPage
	}
}

// This function returns true if a repoUrl is a GitHub release repo
// Arguments:
// - feedUrl: The URL of the RSS feed to check.
func (gh *gitHub) IsReleaseFeed(feedUrl string) bool {
	return gh.isRelRepoRegex.MatchString(feedUrl)
}

package github

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
)

// GitHubStarredFeedBuilder is an interface for retrieving Atom feeds from GitHub
// for all starred repos.
type GitHubStarredFeedBuilder interface {
	GetStarredRepos(ctx context.Context) (map[string]GitHubRepo, error)
	IsGitHubReleasesFeed(feedUrl string) bool
}

// gitHubStarredFeedBuilder is the private implementation of GitHubStarredFeedBuilder.
type gitHubStarredFeedBuilder struct {
	token             string // WARNING: Do not log this value as it is a secret
	client            *http.Client
	nextPageLinkRegex *regexp.Regexp
	isRelRepoRegex    *regexp.Regexp
}

// NewGitHubStarredFeedBuilder creates a new GitHubStarredFeedBuilder instance.
// Arguments:
// - token: The GitHub API token to authenticate with.
// - client: The http client to use for requests (used for mocking).
func NewGitHubStarredFeedBuilder(
	token string,
	client *http.Client,
) GitHubStarredFeedBuilder {
	// This regex is used to find the next page link in the GitHub API response
	nextPageLinkRegex := regexp.MustCompile(`<([^>]+)>; rel="next"`)
	// This regex is used to determine if an RSS feed is a GitHub release feed
	isRelRepoRegex := regexp.MustCompile(`^https://github.com/[\w\.\-]+/[\w\.\-]+/releases\.atom`)
	return &gitHubStarredFeedBuilder{token, client, nextPageLinkRegex, isRelRepoRegex}
}

// This will return all starred repos including the Atom feeds for their releases
// It returns a map of relaseFeedUrl -> GitHubRepo
func (gh *gitHubStarredFeedBuilder) GetStarredRepos(
	ctx context.Context,
) (map[string]GitHubRepo, error) {
	allFeeds := make(map[string]GitHubRepo)
	getUrl := "https://api.github.com/user/starred?per_page=100"
	slog.Debug("Querying GitHub for starred repos", "url", getUrl)

	for {
		ghResponse, err := doApiRequest(ctx, gh.client, getUrl, gh.token, gh.nextPageLinkRegex)
		if err != nil {
			return nil, err
		}

		var repos []GitHubRepo
		if err = json.Unmarshal(ghResponse.data, &repos); err != nil {
			return nil, err
		}

		for _, repo := range repos {
			repo.BuildReleasesFeedURL()
			allFeeds[repo.ReleaseFeedURL] = repo
		}

		// If there is no next page we are done...
		if ghResponse.nextPage == "" {
			return allFeeds, nil
		}

		slog.Debug("Found next page", "url", ghResponse.nextPage)
		getUrl = ghResponse.nextPage
	}
}

// This function returns true if a repoUrl is a GitHub release repo
// Arguments:
// - feedUrl: The URL of the RSS feed to check.
func (gh *gitHubStarredFeedBuilder) IsGitHubReleasesFeed(feedUrl string) bool {
	return gh.isRelRepoRegex.MatchString(feedUrl)
}

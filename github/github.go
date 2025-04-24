package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
)

// This object handles retrieving Atom feeds from GitHub for all starred repos.
type GitHubStarredFeedBuilder struct {
	token             string // WARNING: Do not log this value as it is a secret
	ctx               context.Context
	client            *http.Client
	nextPageLinkRegex *regexp.Regexp
	isRelRepoRegex    *regexp.Regexp
}

// NewGitHubStarredFeedBuilder creates a new GitHubStarredFeedBuilder instance.
// Arguments:
// - token: The GitHub API token to authenticate with.
// - ctx: The context to use for requests.
// - client: The http client to use for requests (used for mocking).
func NewGitHubStarredFeedBuilder(
	token string,
	ctx context.Context,
	client *http.Client,
) *GitHubStarredFeedBuilder {
	// This regex is used to find the next page link in the GitHub API response
	nextPageLinkRegex, _ := regexp.Compile(`<([^>]+)>; rel="next"`)
	// This regex is used to determine if an RSS feed is a GitHub release feed
	isRelRepoRegex, _ := regexp.Compile(`^https://github.com/[\w\.\-]+/[\w\.\-]+/releases\.atom`)
	return &GitHubStarredFeedBuilder{token, ctx, client, nextPageLinkRegex, isRelRepoRegex}
}

// This will return all starred repos including the Atom feeds for their releases
// It returns a map of relaseFeedUrl -> GitHubRepo
func (gh *GitHubStarredFeedBuilder) GetStarredRepos() (map[string]GitHubRepo, error) {
	allFeeds := make(map[string]GitHubRepo)
	getUrl := "http://api.github.com/user/starred?per_page=100"
	slog.Debug("Querying Github for starred repos", "url", getUrl)

	for {
		ghResponse, err := gh.doApiRequest(getUrl)
		if err != nil {
			return nil, err
		}

		var repos []GitHubRepo
		if err = json.Unmarshal(ghResponse.data, &repos); err != nil {
			return nil, err
		}

		for _, repo := range repos {
			repo.BuildReleasesFeedURL()
			allFeeds[repo.FeedUrl] = repo
		}

		// If there is no next page we are done...
		if ghResponse.nextPage == "" {
			return allFeeds, nil
		}

		slog.Debug("Found next page", "url", ghResponse.nextPage)
		getUrl = ghResponse.nextPage
	}
}

// This function returns true if a repoUrl is a Github release repo
// Arguments:
// - feedUrl: The URL of the RSS feed to check.
func (gh *GitHubStarredFeedBuilder) IsGithubReleasesFeed(feedUrl string) bool {
	return gh.isRelRepoRegex.MatchString(feedUrl)
}

func (gh *GitHubStarredFeedBuilder) doApiRequest(url string) (*GithubResponse, error) {
	headers := map[string]string{
		"Authorization":        fmt.Sprintf("Bearer %s", gh.token),
		"X-Github-Api-Version": "2022-11-28",
		"User-Agent":           "github.com/atomicmeganerd/starfeed",
		"Content-Type":         "application/json",
		"Accept":               "application/json",
	}

	// No request will always be valid here so we can ignore the error
	req, _ := http.NewRequestWithContext(gh.ctx, "GET", url, nil)

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := gh.client.Do(req)
	if err != nil {
		slog.Error("Unable to make request to Github", "error", err.Error())
		return nil, err

	}
	defer res.Body.Close() // nolint:all

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned an http error code %d", res.StatusCode)
	}

	ghResponse, err := gh.processGithubResponse(res)
	if err != nil {
		slog.Error("Unable to parse response from Github", "error", err)
		return nil, err
	}

	return ghResponse, nil
}

func (gh *GitHubStarredFeedBuilder) processGithubResponse(r *http.Response) (*GithubResponse, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	linkRaw := r.Header.Get("link")
	links := strings.Split(linkRaw, ",")
	for _, link := range links {
		matches := gh.nextPageLinkRegex.FindStringSubmatch(link)
		if len(matches) == 2 {
			return &GithubResponse{data: data, nextPage: matches[1]}, nil
		}
	}

	return &GithubResponse{data: data}, nil
}

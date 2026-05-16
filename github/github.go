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

	"github.com/atomicmeganerd/starfeed/githost"
)

// gitHubStarredFeedBuilder is the private implementation of GitHubStarredFeedBuilder.
type gitHubStarredFeedBuilder struct {
	gitHost           githost.GitHostConfig
	client            *http.Client
	nextPageLinkRegex *regexp.Regexp
	isRelRepoRegex    *regexp.Regexp
}

// NewGitHubStarredFeedBuilder creates a new GitHubStarredFeedBuilder instance.
// Arguments:
// - cfg: This holds the configuration state this object needs.
// - client: The http client to use for requests (used for mocking).
func NewGitHubStarredFeedBuilder(
	gitHost githost.GitHostConfig,
	client *http.Client,
) githost.GitHost {
	// This regex is used to find the next page link in the GitHub API response
	nextPageLinkRegex := regexp.MustCompile(`<([^>]+)>; rel="next"`)
	// This regex is used to determine if an RSS feed is a GitHub release feed
	isRelRepoRegex := regexp.MustCompile(`^https://github.com/[\w\.\-]+/[\w\.\-]+/releases\.atom`)
	return &gitHubStarredFeedBuilder{gitHost, client, nextPageLinkRegex, isRelRepoRegex}
}

// This will return all starred repos including the Atom feeds for their releases
// It returns a map of releaseFeedUrl -> GitHubRepo
func (gh *gitHubStarredFeedBuilder) GetStarredRepos(
	ctx context.Context,
) (map[string]githost.Repo, error) {
	allFeeds := make(map[string]githost.Repo)
	getUrl := "https://api.github.com/user/starred?per_page=100"
	slog.Debug("Querying GitHub for starred repos", "url", getUrl)

	for {
		ghResponse, err := gh.doApiRequest(ctx, getUrl)
		if err != nil {
			return nil, err
		}

		var repos []GitHubRepo
		if err = json.Unmarshal(ghResponse.data, &repos); err != nil {
			return nil, err
		}

		for _, repo := range repos {
			allFeeds[repo.FeedURL()] = &repo
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
func (gh *gitHubStarredFeedBuilder) IsReleaseFeed(feedUrl string) bool {
	return gh.isRelRepoRegex.MatchString(feedUrl)
}

func (gh *gitHubStarredFeedBuilder) doApiRequest(
	ctx context.Context,
	ghReqUrl string,
) (*GitHubResponse, error) {
	headers := map[string]string{
		"Authorization":        fmt.Sprintf("Bearer %s", gh.gitHost.Token),
		"X-GitHub-Api-Version": "2022-11-28",
		"User-Agent":           "github.com/atomicmeganerd/starfeed",
		"Content-Type":         "application/json",
		"Accept":               "application/json",
	}

	// No request will always be valid here so we can ignore the error
	req, err := http.NewRequestWithContext(ctx, "GET", ghReqUrl, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	res, err := gh.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() // nolint: errcheck

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github returned an http error code %d", res.StatusCode)
	}

	return gh.processGitHubResponse(res)
}

func (gh *gitHubStarredFeedBuilder) processGitHubResponse(
	r *http.Response,
) (*GitHubResponse, error) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	linkRaw := r.Header.Get("Link")
	links := strings.SplitSeq(linkRaw, ",")
	for link := range links {
		matches := gh.nextPageLinkRegex.FindStringSubmatch(link)
		if len(matches) == 2 {
			return &GitHubResponse{data: data, nextPage: matches[1]}, nil
		}
	}

	return &GitHubResponse{data: data}, nil
}

package github

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"regexp"
)

type GitHubSubscribedIssuesFeedBuilder interface {
	GetSubscribedIssues() (map[string]GitHubIssue, error)
}

type gitHubSubscribedIssuesFeedBuilder struct {
	token             string // WARNING: Do not log this value at it is a secret
	ctx               context.Context
	client            *http.Client
	nextPageLinkRegex *regexp.Regexp
}

func NewGitHubSubscribedIssuesFeedBuilder(
	token string,
	ctx context.Context,
	client *http.Client,
) GitHubSubscribedIssuesFeedBuilder {
	nextPageLinkRegex := regexp.MustCompile(`<([^>]+)>; rel="next"`)
	return &gitHubSubscribedIssuesFeedBuilder{token, ctx, client, nextPageLinkRegex}
}

func (gh *gitHubSubscribedIssuesFeedBuilder) GetSubscribedIssues() (map[string]GitHubIssue, error) {
	allFeeds := make(map[string]GitHubIssue)
	getUrl := "https://api.github.com/issues?subscribed&state=all&per_page=50"
	slog.Debug("Querying GitHub for subscribed issues", "url", getUrl)

	for {
		ghResponse, err := doApiRequest(gh.ctx, gh.client, getUrl, gh.token, gh.nextPageLinkRegex)
		if err != nil {
			return nil, err
		}

		var issues []GitHubIssue
		if err = json.Unmarshal(ghResponse.data, &issues); err != nil {
			return nil, err
		}

		for _, issue := range issues {
			allFeeds[issue.Title] = issue
		}

		if ghResponse.nextPage == "" {
			return allFeeds, nil
		}

		slog.Debug("Found next page", "url", ghResponse.nextPage)
		getUrl = ghResponse.nextPage
	}
}

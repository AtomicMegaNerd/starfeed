package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
)

type GitHubSubscribedIssuesFeedBuilder interface {
	GetSubscribedIssues(ctx context.Context) (map[string][]GitHubIssue, error)
}

type gitHubSubscribedIssuesFeedBuilder struct {
	token             string // WARNING: Do not log this value at it is a secret
	client            *http.Client
	nextPageLinkRegex *regexp.Regexp
	ownerRepoRegex    *regexp.Regexp
}

func NewGitHubSubscribedIssuesFeedBuilder(
	token string,
	client *http.Client,
) GitHubSubscribedIssuesFeedBuilder {
	nextPageLinkRegex := regexp.MustCompile(`<([^>]+)>; rel="next"`)
	ownerRepoRegex := regexp.MustCompile(
		`https://api\.github\.com/repos/([a-zA-Z0-9][a-zA-Z0-9-]*)/([a-zA-Z0-9._-]+)`,
	)
	return &gitHubSubscribedIssuesFeedBuilder{
		token,
		client,
		nextPageLinkRegex,
		ownerRepoRegex,
	}
}

// This method should get a list of issues that the current logged in API user
// is subscribed to.
func (gh *gitHubSubscribedIssuesFeedBuilder) GetSubscribedIssues(ctx context.Context) (
	map[string][]GitHubIssue, error,
) {
	allIssuesFeeds := make(map[string][]GitHubIssue)
	getUrl := "https://api.github.com/issues?filter=subscribed&state=all&per_page=100"
	slog.Debug("Querying GitHub for subscribed issues", "url", getUrl)

	for {
		ghResponse, err := doApiRequest(ctx, gh.client, getUrl, gh.token, gh.nextPageLinkRegex)
		if err != nil {
			return nil, err
		}

		var issues []GitHubIssue
		if err = json.Unmarshal(ghResponse.data, &issues); err != nil {
			return nil, err
		}

		for _, issue := range issues {
			slog.Debug("Processing GitHub issue", "issue", issue)
			matches := gh.ownerRepoRegex.FindStringSubmatch(issue.RepositoryURL)
			if len(matches) != 3 {
				slog.Warn(
					"Skipping issue due to invalid repository URL",
					"id", issue.ID,
					"title", issue.Title,
				)
				continue
			}
			issue.Owner = matches[1]
			issue.Repo = matches[2]
			issuesKey := fmt.Sprintf("%s/%s", issue.Owner, issue.Repo)

			// Append the issue to the correct map entry
			allIssuesFeeds[issuesKey] = append(allIssuesFeeds[issuesKey], issue)
		}

		if ghResponse.nextPage == "" {
			return allIssuesFeeds, nil
		}

		slog.Debug("Found next page", "url", ghResponse.nextPage)
		getUrl = ghResponse.nextPage
	}
}

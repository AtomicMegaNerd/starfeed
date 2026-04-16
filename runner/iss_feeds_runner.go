package runner

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/atomicmeganerd/starfeed/github"
)

type IssuesRSSPublisher interface {
	QueryAndPublishFeeds(ctx context.Context) error
}

type issuesRSSPublisher struct {
	ghToken string // WARNING Do not log this value as it is a secret
	client  *http.Client
}

func NewIssuesRSSPublisher(
	ghToken string,
	client *http.Client,
) IssuesRSSPublisher {
	return &issuesRSSPublisher{
		ghToken,
		client,
	}
}

func (i *issuesRSSPublisher) QueryAndPublishFeeds(ctx context.Context) error {
	slog.Info("Starting main issues publish workflow")
	start := time.Now()

	gh := github.NewGitHubSubscribedIssuesFeedBuilder(i.ghToken, ctx, i.client)
	allSubscribedIssues, err := gh.GetSubscribedIssues()
	if err != nil {
		return err
	}

	for key, issues := range allSubscribedIssues {
		slog.Info("Displaying all issues for repo", "repo", key)
		for _, issue := range issues {
			slog.Info("Found Issue", "issue", issue)
		}
	}
	duration := time.Since(start)
	slog.Info("FreshRSS issues feeds synced with GitHub successfully", "duration", duration)

	return nil
}

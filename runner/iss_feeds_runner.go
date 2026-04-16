package runner

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/github"
)

type publishIssuesRunner struct {
	ghToken string // WARNING Do not log this value as it is a secret
	cfg     *config.Config
	client  *http.Client
}

func NewIssuesRSSPublisher(
	cfg *config.Config,
	client *http.Client,
) Runner {
	return &publishIssuesRunner{
		cfg.GitHubToken,
		cfg,
		client,
	}
}

func (i *publishIssuesRunner) Run(ctx context.Context) error {
	if i.cfg.DisableIssueFeedMode {
		slog.Warn("Issues workflow disabled")
		return nil
	}

	slog.Info("Starting main issues publish workflow")
	start := time.Now()

	gh := github.NewGitHubSubscribedIssuesFeedBuilder(i.ghToken, i.client)
	allSubscribedIssues, err := gh.GetSubscribedIssues(ctx)
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

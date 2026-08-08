package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/atomicmeganerd/starfeed/runners"
)

// This function builds our runner objects. We have one shared rssServer but we can have
// multiple git forges so we return one runner per git forge.
func buildRunners(
	ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
	client *http.Client,
) ([]runners.StarfeedRunner, error) {
	// We build a shared RSS server that we publish too. All runners share it.
	rssServerName := cfg.RSSServer.Name
	rssServerLogger := logger.With("rssServer", rssServerName)
	rssServer := rss.NewFreshRSSClient(
		cfg.RSSServer.User, cfg.RSSServer.URL, rssServerLogger, client,
	)
	// Try to authenticate to the shared RSS server
	if err := rssServer.Authenticate(ctx, cfg.RSSServer.Token); err != nil {
		return nil, fmt.Errorf("error authenticating to freshrss %s: %w", rssServerName, err)
	}
	rssServerLogger.Info("Successfully authenticated to RSS Server")

	// For each GitForge in our config let's create a new runner. Each will share a single RSS
	// target but query starred repo from each configured GitForge.
	runnerSlice := make([]runners.StarfeedRunner, len(cfg.GitForges))
	for ix, forgeCfg := range cfg.GitForges {
		forgeName := forgeCfg.Name
		forge := gitforge.NewGitForgeClient(
			forgeCfg.Type,
			forgeCfg.Fqdn,
			forgeCfg.Token,
			logger.With("gitForge", forgeName),
			client,
		)

		// The category we publish in RSS  is always equal to the name of the GitForge
		category := rss.FeedCategory(forgeName)
		syncLogger := logger.With("gitForge", forgeName, "rssServer", rssServerName)
		runner := runners.NewSyncFeedsRunner(
			forge,
			rssServer,
			category,
			syncLogger,
		)

		runnerSlice[ix] = runner
		syncLogger.Info("Successfully registered runner")
	}
	return runnerSlice, nil
}

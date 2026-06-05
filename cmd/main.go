package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/githost"
	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/atomicmeganerd/starfeed/runners"
	"github.com/lmittmann/tint"
	"golang.org/x/sync/errgroup"
)

// This is injected by the CI/CD to tag the binary
var (
	version = "local"
	commit  = ""
)

func main() {
	// The configuration is loaded from the environment
	cfg, err := config.NewConfig(config.OSEnvGetter{})
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: cfg.HTTPTimeout}

	// configure logger
	var logger *slog.Logger
	w := os.Stderr
	if cfg.DebugMode {
		logger = slog.New(
			tint.NewHandler(w, &tint.Options{
				Level:      slog.LevelDebug,
				TimeFormat: time.RFC3339,
			}),
		)
	} else {
		logger = slog.New(
			tint.NewHandler(w, &tint.Options{
				Level:      slog.LevelInfo,
				TimeFormat: time.RFC3339,
			}),
		)
	}

	logger.Info("***********************************************")
	logger.Info(" Welcome to Starfeed", "version", version, "commit", commit)
	logger.Info("***********************************************")

	// Again written by the human:
	// Register signal handling. This will setup a private channel in our ctx object will
	// be closed if one of these signals is received. This is easy to understand...
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Setup our ticker for our timed execution. This will send a time.Time value to the ticker.C
	// every 24 hours.
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Try to authenticate to the target RSS server
	rssServer, err := rss.NewFreshRSS(ctx, cfg.RSSServerConfig, logger, client)
	if err != nil {
		logger.Error("Error authenticating to FreshRSS", "error", err)
		os.Exit(1)
	}
	logger.Info(
		"Successfully authenticated to RSS Server", "rssSerer", cfg.RSSServerConfig.BaseURL,
	)

	// For each GitHost in our config let's create a new runner
	runnerSlice := make([]starfeedRunner, 0)
	for _, gitHostConfig := range cfg.GitHostConfigs {
		gitHost := githost.NewGitHost(gitHostConfig, logger, client)
		releasesRunner := runners.NewSyncFeedsRunner(gitHost, rssServer, logger)
		logger.Info("Successfully registered runner for githost", "name", gitHostConfig.Name)
		runnerSlice = append(runnerSlice, releasesRunner)
	}

	// Always run once...
	if err := executeRunners(ctx, runnerSlice); err != nil {
		logger.Error("Error executing runners", "error", err)
		os.Exit(1)
	}

	if cfg.SingleRunMode {
		logger.Info("Cancelling as we are in single run mode...")
		return
	}

	// The comments below were written by me the human as I try to better understand how Go
	// uses channels and select in this context.
	for {
		select {
		// If the signal handler closes the private channel, the fact the channel was closed will
		// wake up this goroutine and trigger this clause. Done() here is a getter for the private
		// channel that the signal notifier uses behind the scenes. Reading from a closed channel
		// results in no data being returned but all we need here is wake the goroutine and execute
		// the clause.
		case <-ctx.Done():
			logger.Info("Exiting...")
			return
		// ticker.C receives a time.Time value here but we ignore it because our logs will
		// already capture the timestamp when we execute. But it is good to recognize that
		// the ticker channel is sent this data.
		case <-ticker.C:
			if err := executeRunners(ctx, runnerSlice); err != nil {
				logger.Error("Error executing runners", "error", err)
				os.Exit(1)
			}
			logger.Info("Sleeping for 24 hours...")
		}
	}
}

type starfeedRunner interface {
	Run(ctx context.Context) error
}

// Here we execute the runners in parallel...
func executeRunners(ctx context.Context, runners []starfeedRunner) error {
	errGroup, runnerCtx := errgroup.WithContext(ctx)
	for _, runner := range runners {
		errGroup.Go(func() error {
			return runner.Run(runnerCtx)
		})
	}
	return errGroup.Wait()
}

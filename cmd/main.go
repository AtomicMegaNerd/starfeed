package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atomicmeganerd/starfeed/atom"
	"github.com/atomicmeganerd/starfeed/config"
	"github.com/atomicmeganerd/starfeed/githost"
	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/atomicmeganerd/starfeed/runner"
	"github.com/lmittmann/tint"
)

func main() {
	slog.Info("***********************************************")
	slog.Info(" Welcome to Starfeed")
	slog.Info("***********************************************")

	// The configuration is loaded from the environment
	cfg, err := config.NewConfig(config.OSEnvGetter{})
	if err != nil {
		slog.Error("Failed to load configuration", "error", err.Error())
		os.Exit(1)
	}

	client := &http.Client{Timeout: cfg.HTTPTimeout}

	// configure logger
	w := os.Stderr
	if cfg.DebugMode {
		slog.SetDefault(slog.New(
			tint.NewHandler(w, &tint.Options{
				Level:      slog.LevelDebug,
				TimeFormat: time.RFC3339,
			}),
		))
	} else {
		slog.SetDefault(slog.New(
			tint.NewHandler(w, &tint.Options{
				Level:      slog.LevelInfo,
				TimeFormat: time.RFC3339,
			}),
		))
	}

	// Again written by the human:
	// Register signal handling. This will setup a private channel in our ctx object will
	// be closed if one of these signals is received. This is easy to understand...
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Setup our ticker for our timed execution. This will send a time.Time value to the ticker.C
	// every 24 hours.
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	runners := make([]runner.Runner, 0)
	feedChecker := atom.NewAtomFeedChecker(client)
	rssServer := rss.NewFreshRSSFeedManager(cfg.RSSServerConfig, client)
	slog.Info("Successfully registered RSS server...", "URL", cfg.RSSServerConfig.BaseURL)

	// For each GitHost in our config let's create a new runner
	for _, gitHostConfig := range cfg.GitHostConfigs {
		gitHost, err := githost.NewGitHost(gitHostConfig, client)
		slog.Info("Successfully registered git host", "name", gitHostConfig.Name)
		if err != nil {
			slog.Error("Cannot configure git host...", "error", err)
			os.Exit(1)
		}
		releasesRunner := runner.NewPublishReleasesRunner(gitHost, rssServer, feedChecker)
		runners = append(runners, releasesRunner)
	}

	// Always run once...
	if err := executeRunners(ctx, runners); err != nil {
		slog.Error("Error executing runers", "error", err)
		os.Exit(1)
	}

	if cfg.SingleRunMode {
		slog.Info("Cancelling as we are in single run mode...")
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
			slog.Info("Exiting...")
			return
		// ticker.C receives a time.Time value here but we ignore it because our logs will
		// already capture the timestamp when we execute. But it is good to recognize that
		// the ticker channel is sent this data.
		case <-ticker.C:
			if err := executeRunners(ctx, runners); err != nil {
				slog.Error("Error executing runers", "error", err)
				os.Exit(1)
			}
			slog.Info("Sleeping for 24 hours...")
		}
	}
}

func executeRunners(ctx context.Context, runners []runner.Runner) error {
	for _, runner := range runners {
		if err := runner.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

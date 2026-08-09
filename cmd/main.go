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
	"github.com/atomicmeganerd/starfeed/gitforge"
	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/atomicmeganerd/starfeed/runners"
)

// This is injected by the CI/CD to tag the binary
var (
	version = "local"
	commit  = ""
)

func main() {
	// Return an error to the operating system if run returns an error
	if err := run(); err != nil {
		os.Exit(1)
	}
}

// Return errors so we can return an error to the OS in main
func run() error {
	// The configuration is loaded from the environment
	cfg, err := config.NewConfig(config.ConfigLoader{})
	if err != nil {
		slog.Default().Error("Error loading configuration", "error", err)
		return err
	}

	logger := buildLogger(cfg.Debug)

	logger.Info("***********************************************")
	logger.Info(" Welcome to Starfeed", "version", version, "commit", commit)
	logger.Info("***********************************************")
	logger.Debug("Debug mode enabled")

	// Register signal handling. This will setup a private channel in our ctx object which will
	// be closed if one of these signals is received. This is easy to understand...
	// NOTE: the channel in ctx is one-shot and is a synchronization channel (meaning no actual
	// data is sent).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	client := &http.Client{Timeout: 60 * time.Second}

	rssServer := rss.NewFreshRSS(
		cfg.RSSServer.User, cfg.RSSServer.URL, logger, client,
	)
	if err := rssServer.Authenticate(ctx, cfg.RSSServer.Token); err != nil {
		logger.Error("Error authenticatin to RSS", "error", err)
		return err
	}

	// Setup our ticker for our timed execution. This will send a time.Time value to the ticker.C
	// is set by the interval setting in the Config.
	// NOTE: This is a bounded (size 1) async channel
	ticker := time.NewTicker(cfg.Interval())
	defer ticker.Stop()

	gitForges := make([]gitforge.GitForge, len(cfg.GitForges))
	for ix, gitForgeCfg := range cfg.GitForges {
		gitForges[ix] = gitforge.NewGitForgeClient(
			gitForgeCfg.Name,
			gitForgeCfg.Type,
			gitForgeCfg.Fqdn,
			gitForgeCfg.Token,
			client,
		)
	}

	runner := runners.NewRunner(ctx, stop, *rssServer, gitForges, logger)
	runner.Init(ctx)

	if cfg.SingleRun {
		logger.Info("Cancelling as we are in single run mode...")
		return nil
	}

	for {
		// Select will block until one of the two signals are received. The goroutine is parked
		// until one of the below channels sends a message.
		select {
		// If the signal handler closes the private channel, the fact the channel was closed will
		// wake up this goroutine and trigger this clause. Done() here is a getter for the private
		// channel that the signal notifier uses behind the scenes. Reading from a closed channel
		// results in no data being returned but all we need here is wake the goroutine and execute
		// the clause.
		case <-ctx.Done():
			logger.Info("Exiting...")
			return nil
			// ticker.C receives a time.Time value here but we ignore it because our logs will
			// already capture the timestamp when we execute. But it is good to recognize that
			// the ticker channel is sent this data.
		case t := <-ticker.C:
			runner.RunChan <- struct{}{}
			logger.Info("Sleeping...", "nextRun", t.Add(cfg.Interval()))
		}
	}
}

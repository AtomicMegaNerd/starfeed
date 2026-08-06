package main

import (
	"context"
	"fmt"
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
	"github.com/lmittmann/tint"
	"golang.org/x/sync/errgroup"
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

	client := &http.Client{Timeout: 60 * time.Second}
	logger := getLogger(cfg.Debug)

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

	// Setup our ticker for our timed execution. This will send a time.Time value to the ticker.C
	// every 24 hours.
	// NOTE: This is a bounded (size 1) async channel
	ticker := time.NewTicker(cfg.Interval())
	defer ticker.Stop()

	runnerSlice, err := buildRunners(ctx, cfg, logger, client)
	if err != nil {
		logger.Error("Error building runners", "error", err)
		return err
	}

	// We always want to run on startup, and if we are in SingleRun mode we will terminate
	// the app after running the workflow once. SingleRun is useful for development and testing.
	if err := executeRunners(ctx, runnerSlice); err != nil {
		logger.Error("Error executing runners", "error", err)
		return err
	}
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
			if err := executeRunners(ctx, runnerSlice); err != nil {
				logger.Error("Error executing runners", "error", err)
				return err
			}
			logger.Info("Sleeping...", "nextRun", t.Add(24*time.Hour))
		}
	}
}

// Defining an interface here for any Runner that we want to add to our runnerSlice
type starfeedRunner interface {
	Run(ctx context.Context) error
}

// This function builds our runner objects. We have one shared rssServer but we can have
// multiple git forges so we return one runner per git forge.
func buildRunners(
	ctx context.Context,
	cfg config.Config,
	logger *slog.Logger,
	client *http.Client,
) ([]starfeedRunner, error) {
	rssServer := rss.NewFreshRSS(
		cfg.RSSServer.Name, cfg.RSSServer.User, cfg.RSSServer.URL, logger, client,
	)
	// Try to authenticate to the target RSS server
	if err := rssServer.Authenticate(ctx, cfg.RSSServer.Token); err != nil {
		return nil, fmt.Errorf("error authenticating to freshrss: %w", err)
	}
	logger.Info(
		"Successfully authenticated to RSS Server", "rssServer", cfg.RSSServer.URL,
	)

	runnerSlice := make([]starfeedRunner, 0)
	// For each GitForge in our config let's create a new runner
	for _, forgeCfg := range cfg.GitForges {
		forge := gitforge.NewGitForge(
			forgeCfg.Type, forgeCfg.Name, forgeCfg.Fqdn, forgeCfg.Token, logger, client,
		)
		runner := runners.NewSyncFeedsRunner(forge, rssServer, logger)
		logger.Info("Successfully registered runner for gitForge", "name", forgeCfg.Name)
		runnerSlice = append(runnerSlice, runner)
	}
	return runnerSlice, nil
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

// This configures the logger for our application setting the level to debug
// if specified.
func getLogger(debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	}
	return slog.New(
		tint.NewTextHandler(
			os.Stderr,
			&tint.Options{Level: level, TimeFormat: time.RFC3339},
		),
	)
}

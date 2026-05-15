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
	"github.com/atomicmeganerd/starfeed/rss"
	"github.com/atomicmeganerd/starfeed/runner"
	"github.com/lmittmann/tint"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	defer cancel()

	slog.Info("***********************************************")
	slog.Info(" Welcome to Starfeed")
	slog.Info("***********************************************")

	cfg, err := config.NewConfig(config.OSEnvGetter{})
	if err != nil {
		slog.Error("Failed to load configuration", "error", err.Error())
		os.Exit(1)
	}

	// configure logger
	w := os.Stderr
	if cfg.DebugMode {
		slog.SetDefault(slog.New(
			tint.NewHandler(w, &tint.Options{
				Level:      slog.LevelDebug,
				TimeFormat: time.Kitchen,
			}),
		))
	} else {
		slog.SetDefault(slog.New(
			tint.NewHandler(w, &tint.Options{
				Level:      slog.LevelInfo,
				TimeFormat: time.Kitchen,
			}),
		))
	}

	// In this case both os.Interrupt and syscall.SIGTERM are signals.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		slog.Warn("Received interrupt signal, shutting down...")
		cancel()
	}()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Let's run the RSS server in a separate Go routine
	rss := rss.NewRSSServer(cfg)
	g.Go(func() error {
		return rss.Start(ctx)
	})

	releasesRunner := runner.NewPublishReleasesRunner(
		cfg,
		&http.Client{Timeout: cfg.HTTPTimeout},
	)

	issuesRunner := runner.PublishIssuesRunner(
		cfg,
		&http.Client{Timeout: cfg.HTTPTimeout},
	)

	runners := []runner.Runner{
		releasesRunner, issuesRunner,
	}

	g.Go(func() error {
		// Always run once.... if we are in SingleRunMode we will return and terminate the program
		if err := executeRunners(ctx, runners); err != nil {
			slog.Error("Error executing runners", "error", err)
			return err
		}

		// Cancel further execution if we are in SingleRunMode
		if cfg.SingleRunMode {
			slog.Info("Cancelling as we are in single run mode...")
			cancel()
		}

		for {
			select {
			case <-ctx.Done():
				slog.Info("Exiting...")
				return nil
			case <-ticker.C:
				if err := executeRunners(ctx, runners); err != nil {
					slog.Error("Error executing runners", "error", err)
					return err
				}
				slog.Info("Sleeping for 24 hours...")
			}
		}
	})

	if err := g.Wait(); err != nil {
		slog.Error("Fatal error resulting in shutdown...", "error", err)
		os.Exit(1)
	}
}

// This function executes all of the runners sequentially. If SingleRunMode is set it returns
// true, otherwise false. It will also return any errors.
func executeRunners(
	ctx context.Context,
	runners []runner.Runner,
) error {
	for _, r := range runners {
		if err := r.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

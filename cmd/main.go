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

	// Register signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Setup cancel function for SingleRunMode
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup our ticker for our timed execution
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Setup our primary error group
	g, ctx := errgroup.WithContext(ctx)

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
		// Always run once...
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

	// NOTE: By the time g.Wait() returns if there is an error the RSS server will have
	// gracefully shutdown because we used errgroup.WithContext()
	if err := g.Wait(); err != nil {
		slog.Error("Fatal error resulting in shutdown...", "error", err)
		os.Exit(1)
	}
}

// This function executes all of the runners in parallel using its own errgroup. It returns the
// first error it finds
func executeRunners(
	ctx context.Context,
	runners []runner.Runner,
) error {
	g, ctx := errgroup.WithContext(ctx)
	for _, r := range runners {
		g.Go(func() error { return r.Run(ctx) })
	}
	return g.Wait()
}

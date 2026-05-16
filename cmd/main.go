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
	"github.com/atomicmeganerd/starfeed/runner"
	"github.com/lmittmann/tint"
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

	// Again written by the human:
	// Register signal handling. This will setup a private channel in our ctx object will
	// be closed if one of these signals is received. This is easy to understand...
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Setup our ticker for our timed execution. This will send a time.Time value to the ticker.C
	// every 24 hours.
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	releasesRunner := runner.NewPublishReleasesRunner(
		cfg,
		&http.Client{Timeout: cfg.HTTPTimeout},
	)

	// Always run once...
	if err := releasesRunner.Run(ctx); err != nil {
		slog.Error("Error executing runners", "error", err)
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
			if err := releasesRunner.Run(ctx); err != nil {
				slog.Error("Error executing runners", "error", err)
			}
			slog.Info("Sleeping for 24 hours...")
		}
	}
}

package rss

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/atomicmeganerd/starfeed/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RSSServer interface {
	Start(ctx context.Context) error
}

type rssServer struct {
	cfg *config.Config
}

func NewRSSServer(cfg *config.Config) RSSServer {
	return &rssServer{
		cfg: cfg,
	}
}

// This function starts the RSS server
func (rs *rssServer) Start(ctx context.Context) error {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	// This middleware recovers from panics and writes a 500 if there is one.
	r.Use(middleware.Recoverer)
	// NOTE: This middleware logs the IP address of the requestor. This is crucial for our
	// rate limiter
	r.Use(middleware.RealIP)
	// This middleware adds a request ID to each request.
	r.Use(middleware.RequestID)
	// What a great way to set timeout!
	r.Use(middleware.Timeout(rs.cfg.HTTPTimeout))

	r.Route("/feeds", func(r chi.Router) {
		r.Route("/{feedName}", func(r chi.Router) {
			r.Get("/", nil)
		})
	})

	srv := &http.Server{
		Addr:    rs.cfg.RSSServerAddress,
		Handler: r,
	}

	slog.Info("RSS server listening...", "address", rs.cfg.RSSServerAddress)

	go func() { _ = srv.ListenAndServe() }()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Unexpected error on shutodwn...", "error", err)
		return err
	}

	slog.Info("Graceful shudown of RSS server...")
	return nil
}

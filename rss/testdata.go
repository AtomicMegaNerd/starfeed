package rss

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/atomicmeganerd/starfeed/config"
)

var MockValidRSSServer = func(
	ctx context.Context, client *http.Client, logger *slog.Logger,
) FreshRSS {
	rss, _ := NewFreshRSS(
		ctx,
		config.MockValidFreshRSSConfig,
		logger,
		client,
	)
	return rss
}

var MockValidRSSEnabledServer = func(
	ctx context.Context, client *http.Client, logger *slog.Logger,
) (FreshRSS, error) {
	rss, err := NewFreshRSS(
		ctx,
		config.MockValidFreshRSSEnabledConfig,
		logger,
		client,
	)
	return rss, err
}

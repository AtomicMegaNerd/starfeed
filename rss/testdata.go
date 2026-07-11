package rss

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/atomicmeganerd/starfeed/testutils"
)

var (
	MockValidFreshRSSConfig = RSSServerConfig{
		Type:    FreshRSSType,
		BaseURL: testutils.FreshRSSURL,
		User:    testutils.FreshRSSUser,
		Enabled: false,
	}

	MockValidFreshRSSEnabledConfig = RSSServerConfig{
		Type:    FreshRSSType,
		BaseURL: testutils.FreshRSSURL,
		User:    testutils.FreshRSSUser,
		Enabled: true,
	}
)

func MockValidRSSServer(
	ctx context.Context, client *http.Client, logger *slog.Logger,
) FreshRSS {
	rss, _ := NewFreshRSS(
		ctx,
		MockValidFreshRSSConfig,
		logger,
		client,
	)
	return rss
}

func MockValidRSSEnabledServer(
	ctx context.Context, client *http.Client, logger *slog.Logger,
) (FreshRSS, error) {
	rss, err := NewFreshRSS(
		ctx,
		MockValidFreshRSSEnabledConfig,
		logger,
		client,
	)
	return rss, err
}

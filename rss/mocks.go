package rss

import (
	"log/slog"
	"net/http"

	"github.com/atomicmeganerd/starfeed/mocks"
)

var MockValidRSSServer = func(client *http.Client, logger *slog.Logger) FreshRSS {
	return FreshRSS{
		isEnabled: true,
		rssType:   mocks.FreshRSSType,
		baseURL:   mocks.FreshRSSURL,
		user:      mocks.FreshRSSUser,
		logger:    logger,
		client:    client,
	}
}

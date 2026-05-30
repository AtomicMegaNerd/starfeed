package rss

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/atomicmeganerd/starfeed/mocks"
)

var MockValidRSSServer = func(client *http.Client) *FreshRSS {
	return &FreshRSS{
		IsEnabled: true,
		rssType:   mocks.FreshRSSType,
		baseURL:   mocks.FreshRSSURL,
		user:      mocks.FreshRSSUser,
		token:     mocks.FreshRSSToken,
		logger:    slog.New(slog.NewTextHandler(io.Discard, nil)),
		client:    client,
	}
}

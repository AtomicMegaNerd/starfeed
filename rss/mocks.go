package rss

import (
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
		logger:    mocks.TestLogger(),
		client:    client,
	}
}

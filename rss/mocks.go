package rss

import (
	"net/http"

	"github.com/atomicmeganerd/starfeed/mocks"
)

var MockValidRSSServer = func(client *http.Client) RSSServer {
	return &freshRSS{
		rssType: mocks.FreshRSSType,
		baseURL: mocks.FreshRSSURL,
		user:    mocks.FreshRSSUser,
		token:   mocks.FreshRSSToken,
		enabled: true,
		client:  client,
	}
}

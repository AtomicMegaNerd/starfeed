package rss

import (
	"net/http"

	"github.com/atomicmeganerd/starfeed/mocks"
)

var MockValidRSSServer = func(client *http.Client) *FreshRSS {
	return &FreshRSS{
		isEnabled: true,
		rssType:   mocks.FreshRSSType,
		baseURL:   mocks.FreshRSSURL,
		user:      mocks.FreshRSSUser,
		logger:    mocks.TestLogger(),
		client:    client,
	}
}

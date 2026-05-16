package rss

import "github.com/atomicmeganerd/starfeed/mocks"

var MockValidRSSServer = RSSServerConfig{
	Type:    mocks.FreshRSSType,
	BaseURL: mocks.FreshRSSURL,
	User:    mocks.FreshRSSUser,
	Token:   mocks.FreshRSSToken,
}

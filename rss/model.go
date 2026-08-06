package rss

import "github.com/atomicmeganerd/starfeed/common"

type RSSFeedSet map[common.FeedURL]struct{}

type FreshRSSAddFeedResponse struct {
	NumResults int    `json:"numResults"`
	Query      string `json:"query"`
	StreamId   string `json:"streamId"`
	StreamName string `json:"streamName"`
}

type RSSFeedList struct {
	Feeds []RSSFeed `json:"subscriptions"`
}

type RSSFeed struct {
	URL        common.FeedURL    `json:"url"`
	Categories []RSSFeedCategory `json:"categories"`
}

type RSSFeedCategory struct {
	Label string `json:"label"`
}

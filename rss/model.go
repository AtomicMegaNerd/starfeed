package rss

import (
	"github.com/atomicmeganerd/starfeed/common"
)

type FeedName string

func (n FeedName) String() string {
	return string(n)
}

type FeedCategory string

func (c FeedCategory) String() string {
	return string(c)
}

type AddFeedRequest struct {
	Name     FeedName
	URL      common.FeedURL
	Category FeedCategory
}

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
	Label FeedCategory `json:"label"`
}

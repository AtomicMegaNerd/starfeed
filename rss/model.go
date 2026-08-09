package rss

import (
	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
)

type AddFeedRequest struct {
	Name     gitforge.GitRepoName
	URL      common.FeedURL
	Category gitforge.GitForgeName
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
	Label gitforge.GitForgeName `json:"label"`
}

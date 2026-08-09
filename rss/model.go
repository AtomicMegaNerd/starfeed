package rss

import (
	"github.com/atomicmeganerd/starfeed/common"
	"github.com/atomicmeganerd/starfeed/gitforge"
)

type SubscribeRequest struct {
	Name     gitforge.RepoName
	URL      common.FeedURL
	Category gitforge.ForgeName
}

type SubscribeResponse struct {
	NumResults int    `json:"numResults"`
	Query      string `json:"query"`
	StreamId   string `json:"streamId"`
	StreamName string `json:"streamName"`
}

type FeedList struct {
	Feeds []RSSFeed `json:"subscriptions"`
}

type RSSFeed struct {
	URL        common.FeedURL `json:"url"`
	Categories []FeedCategory `json:"categories"`
}

type FeedCategory struct {
	Label gitforge.ForgeName `json:"label"`
}

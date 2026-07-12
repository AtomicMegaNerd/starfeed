package rss

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
	URL string `json:"url"`
}

// This type both holds and validates the config for the RSS Server
type RSSServerConfig struct {
	Name  string `validate:"required,oneof=freshrss"`
	URL   string `validate:"required,url"`
	User  string `validate:"required,min=3"`
	Token string `validate:"required,min=10"`
}

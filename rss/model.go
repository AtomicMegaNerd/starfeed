package rss

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

type RSSServerConfig struct {
	Type  string `validate:"required,oneof=freshrss"`
	URL   string `validate:"required,url"`
	User  string `validate:"required,min=3"`
	Token string `validate:"required,min=10"`
}

func ParseRSSServerConfigFromCSV(csvLine string) (*RSSServerConfig, error) {
	validate := validator.New()
	parts := strings.SplitN(csvLine, ",", 4)
	if len(parts) != 4 {
		return nil, errors.New("RSSServerConfig invalid")
	}

	rssConfig := &RSSServerConfig{
		Type:  strings.TrimSpace(parts[0]),
		URL:   strings.TrimSpace(parts[1]),
		User:  strings.TrimSpace(parts[2]),
		Token: strings.TrimSpace(parts[3]),
	}

	return rssConfig, validate.Struct(rssConfig)
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
	Url string `json:"url"`
}

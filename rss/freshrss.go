package rss

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/atomicmeganerd/starfeed/common"
)

type FreshRSS struct {
	cfg            RSSServerConfig
	addUrl         string
	addCategoryUrl string
	editUrl        string
	feeds          FeedSet
	logger         *slog.Logger
	headers        http.Header
	client         *http.Client
}

func NewFreshRSS(
	cfg RSSServerConfig,
	logger *slog.Logger,
	client *http.Client,
) *FreshRSS {
	logger = logger.With("rssServer", cfg.Name)

	headers := http.Header{}
	headers.Set("Content-type", "application/x-www-form-urlencoded")

	return &FreshRSS{
		cfg:    cfg,
		logger: logger,
		addUrl: fmt.Sprintf(
			"%s/api/greader.php/reader/api/0/subscription/quickadd",
			cfg.BaseURL,
		),
		addCategoryUrl: fmt.Sprintf(
			"%s/api/greader.php/reader/api/0/subscription/edit",
			cfg.BaseURL,
		),
		editUrl: fmt.Sprintf(
			"%s/api/greader.php/reader/api/0/subscription/edit",
			cfg.BaseURL,
		),
		headers: headers,
		client:  client,
	}
}

// This function will authenticate to FreshRSS.
func (f *FreshRSS) Authenticate(
	ctx context.Context,
) error {
	reqURL := fmt.Sprintf("%s/api/greader.php/accounts/ClientLogin", f.cfg.BaseURL)
	f.logger.Debug("Authenticating with FreshRSS", "url", reqURL)
	formData := []byte(
		url.Values{
			"Email":  {f.cfg.User},
			"Passwd": {f.cfg.Token},
		}.Encode(),
	)
	data, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, reqURL, formData, f.headers, f.client,
	)
	if err != nil {
		return fmt.Errorf("error authenticating to RSS Server: %w, url: %s", err, reqURL)
	}

	var authToken string
	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		if after, ok := strings.CutPrefix(line, "Auth="); ok {
			authToken = after
		}
	}

	f.headers.Set("Authorization", fmt.Sprintf("GoogleLogin auth=%s", authToken))
	return nil
}

func (f *FreshRSS) LoadFeeds(
	ctx context.Context,
) error {
	getURL := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/list?output=json", f.cfg.BaseURL,
	)

	// Perform the request
	res, _, err := common.DoAPIRequest(ctx, http.MethodGet, getURL, nil, f.headers, f.client)
	if err != nil {
		return err
	}

	// Parse the response
	var feeds RSSFeedList
	if err = json.Unmarshal(res, &feeds); err != nil {
		return err
	}

	for _, feed := range feeds.Feeds {
		f.feeds[feed.URL] = struct{}{}
	}

	return nil
}

func (f *FreshRSS) AddFeed(
	ctx context.Context,
	feedURL, name, category string,
) error {

	// Check if feed exists already
	if _, exists := f.feeds[feedURL]; exists {
		f.logger.Debug("Not adding feed as it is already in RSS", "feed", name)
		return nil
	}

	formData := url.Values{
		"quickadd": {feedURL},
	}

	res, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, f.addUrl, []byte(formData.Encode()), f.headers, f.client,
	)
	if err != nil {
		return err
	}

	// Parse the response
	var feedResponse FreshRSSAddFeedResponse
	if err = json.Unmarshal(res, &feedResponse); err != nil {
		return err
	}

	// Add the sub to the category
	if err = f.addFeedToCategory(ctx, feedResponse.StreamId, name, category); err != nil {
		return err
	}

	f.logger.Debug("Successfully added feed to FreshRSS", "feed", feedURL)
	return nil
}

func (f *FreshRSS) Feeds() map[string]struct{} {
	return f.feeds
}

func (f *FreshRSS) RemoveFeed(ctx context.Context, feedURL string) error {

	formData := url.Values{
		"ac": {"unsubscribe"},
		"s":  {fmt.Sprintf("feed/%s", feedURL)},
	}

	// We do not care about the response
	if _, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, f.editUrl, []byte(formData.Encode()), f.headers, f.client,
	); err != nil {
		return err
	}

	return nil
}

func (f *FreshRSS) Name() string {
	return f.cfg.Name
}

func (f *FreshRSS) addFeedToCategory(
	ctx context.Context,
	streamId, name, category string,
) error {
	formData := url.Values{
		"ac": {"edit"},
		"s":  {streamId},
		"t":  {name},
		"a":  {fmt.Sprintf("user/%s/label/%s", f.cfg.User, category)},
	}

	if _, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, f.addCategoryUrl, []byte(formData.Encode()), f.headers, f.client,
	); err != nil {
		return err
	}

	return nil
}

package rss

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/atomicmeganerd/starfeed/common"
)

type FreshRSS struct {
	cfg     RSSServerConfig
	feeds   map[string]struct{}
	logger  *slog.Logger
	headers http.Header
	client  *http.Client
	// Because we share this instance with multiple runners we have to
	// protect the feeds set
	mtx sync.RWMutex
}

func NewFreshRSS(
	cfg RSSServerConfig,
	logger *slog.Logger,
	client *http.Client,
) *FreshRSS {
	headers := http.Header{}
	return &FreshRSS{
		cfg:     cfg,
		logger:  logger.With("rssServer", cfg.Name),
		feeds:   make(map[string]struct{}, 0),
		headers: headers,
		client:  client,
	}
}

// This function will authenticate to FreshRSS.
func (f *FreshRSS) Authenticate(
	ctx context.Context,
) error {
	reqURL := fmt.Sprintf("%s/api/greader.php/accounts/ClientLogin", f.cfg.URL)
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
		f.logger.Debug("Error authenticating", "error", err)
		return fmt.Errorf("error authenticating to RSS Server: %w, url: %s", err, reqURL)
	}

	var authToken string
	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		if after, ok := strings.CutPrefix(line, "Auth="); ok {
			authToken = after
		}
	}

	if authToken == "" {
		return errors.New("failed to parse authToken")
	}

	// We can set all required headers after we authenticate
	f.headers.Set("Authorization", fmt.Sprintf("GoogleLogin auth=%s", authToken))
	f.headers.Set("Content-type", "application/x-www-form-urlencoded")
	return nil
}

func (f *FreshRSS) LoadFeeds(
	ctx context.Context,
) error {
	// Clear the feeds set before reloading...
	newFeeds := make(map[string]struct{}, 0)
	loadUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/list?output=json",
		f.cfg.URL,
	)
	res, _, err := common.DoAPIRequest(ctx, http.MethodGet, loadUrl, nil, f.headers, f.client)
	if err != nil {
		return err
	}

	// Parse the response
	feeds := &RSSFeedList{}
	if err = json.Unmarshal(res, &feeds); err != nil {
		return err
	}

	for _, feed := range feeds.Feeds {
		newFeeds[feed.URL] = struct{}{}
	}

	f.mtx.Lock()
	defer f.mtx.Unlock()
	f.feeds = newFeeds
	return nil
}

func (f *FreshRSS) AddFeed(
	ctx context.Context,
	feedURL, name, category string,
) error {
	// Check if feed exists already
	f.mtx.RLock()
	_, exists := f.feeds[feedURL]
	f.mtx.RUnlock()

	if exists {
		f.logger.Debug("Not adding feed as it is already in RSS", "feed", name)
		return nil
	}

	addUrl := fmt.Sprintf("%s/api/greader.php/reader/api/0/subscription/quickadd", f.cfg.URL)
	formData := url.Values{
		"quickadd": {feedURL},
	}
	res, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, addUrl, []byte(formData.Encode()), f.headers, f.client,
	)
	if err != nil {
		return err
	}

	feedResponse := &FreshRSSAddFeedResponse{}
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

func (f *FreshRSS) RemoveFeed(ctx context.Context, feedURL string) error {
	editUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit",
		f.cfg.URL,
	)
	formData := url.Values{
		"ac": {"unsubscribe"},
		"s":  {fmt.Sprintf("feed/%s", feedURL)},
	}

	// We do not care about the response
	if _, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, editUrl, []byte(formData.Encode()), f.headers, f.client,
	); err != nil {
		return err
	}

	return nil
}

func (f *FreshRSS) Feeds() map[string]struct{} {
	f.mtx.RLock()
	defer f.mtx.RUnlock()
	feedsCopy := maps.Clone(f.feeds)
	return feedsCopy
}

func (f *FreshRSS) Name() string {
	return f.cfg.Name
}

func (f *FreshRSS) addFeedToCategory(
	ctx context.Context,
	streamId, name, category string,
) error {
	addCategoryUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit",
		f.cfg.URL,
	)
	formData := url.Values{
		"ac": {"edit"},
		"s":  {streamId},
		"t":  {name},
		"a":  {fmt.Sprintf("user/%s/label/%s", f.cfg.User, category)},
	}

	if _, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, addCategoryUrl, []byte(formData.Encode()), f.headers, f.client,
	); err != nil {
		return err
	}
	return nil
}

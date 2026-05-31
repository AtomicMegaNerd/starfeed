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
	"github.com/atomicmeganerd/starfeed/config"
)

// The FreshRSS is a private struct that implements FreshRSSFeedManager.
type FreshRSS struct {
	rssType   string
	baseURL   string
	user      string
	isEnabled bool
	logger    *slog.Logger
	headers   http.Header
	client    *http.Client
}

// NewFreshRSS creates a new FreshRSSFeedManager instance.
// Arguments:
// - cfg: This holds the configuration state this object needs.
// - client: The http client to use for requests (used for mocking).
func NewFreshRSS(
	ctx context.Context,
	rssConfig config.RSSServerConfig,
	logger *slog.Logger,
	client *http.Client,
) (FreshRSS, error) {
	logger = logger.With("rssServer", rssConfig.Type)
	headers := http.Header{}
	headers.Set("Content-type", "application/x-www-form-urlencoded")

	// Only authenticate if the RSS server is enabled
	if rssConfig.Enabled {
		authToken, err := authenticate(ctx, rssConfig, headers, logger, client)
		if err != nil {
			return FreshRSS{}, err
		}
		headers.Set("Authorization", fmt.Sprintf("GoogleLogin auth=%s", authToken))
	}

	return FreshRSS{
		rssType:   rssConfig.Type,
		baseURL:   rssConfig.BaseURL,
		user:      rssConfig.User,
		isEnabled: rssConfig.Enabled,
		logger:    logger,
		headers:   headers,
		client:    client,
	}, nil
}

// AddFeed adds a feed to the FreshRSS instance.
// Arguments:
// - feedURL: The URL of the feed to add.
// - name: The name of the feed.
// - category: The category to add the feed to.
func (f FreshRSS) AddFeed(
	ctx context.Context,
	feedURL, name, category string,
) error {
	if !f.isEnabled {
		return nil
	}

	addURL := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/quickadd", f.baseURL,
	)
	formData := url.Values{
		"quickadd": {feedURL},
	}

	f.logger.Debug("Adding feed to FreshRSS", "url", addURL)
	res, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, addURL, []byte(formData.Encode()), f.headers, f.client,
	)
	if err != nil {
		return err
	}

	// Parse the response
	var resData FreshRSSAddFeedResponse
	if err = json.Unmarshal(res, &resData); err != nil {
		return err
	}

	// Add the sub to the category
	if err = f.addFeedToCategory(ctx, resData.StreamId, name, category); err != nil {
		return err
	}

	f.logger.Debug("Successfully added feed to FreshRSS", "feed", feedURL)
	return nil
}

// GetExistingFeeds gets the existing feeds from the FreshRSS instance.
func (f FreshRSS) GetExistingFeeds(ctx context.Context) (map[string]struct{}, error) {
	if !f.isEnabled {
		return nil, nil
	}

	getURL := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/list?output=json", f.baseURL,
	)

	// Perform the request
	res, _, err := common.DoAPIRequest(ctx, http.MethodGet, getURL, nil, f.headers, f.client)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var feeds RSSFeedList
	if err = json.Unmarshal(res, &feeds); err != nil {
		return nil, err
	}

	// NOTE: In Go map[T]struct{} is the idiomatic way to make a set as struct{} is 0-bytes
	feedSet := make(map[string]struct{})
	for _, feed := range feeds.Feeds {
		feedSet[feed.URL] = struct{}{}
	}
	return feedSet, nil
}

// RemoveFeed removes a feed from the FreshRSS instance.
// Arguments:
// - context
// - feedURL: The URL of the feed to remove.
func (f FreshRSS) RemoveFeed(ctx context.Context, feedURL string) error {
	if !f.isEnabled {
		return nil
	}

	rmURL := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit", f.baseURL,
	)

	formData := url.Values{
		"ac": {"unsubscribe"},
		"s":  {fmt.Sprintf("feed/%s", feedURL)},
	}

	// We do not care about the response
	_, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, rmURL, []byte(formData.Encode()), f.headers, f.client,
	)
	if err != nil {
		return err
	}

	return nil
}

func (f *FreshRSS) addFeedToCategory(
	ctx context.Context,
	streamId, name, category string,
) error {
	if !f.isEnabled {
		return nil
	}

	addURL := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit", f.baseURL,
	)

	formData := url.Values{
		"ac": {"edit"},
		"s":  {streamId},
		"t":  {name},
		"a":  {fmt.Sprintf("user/%s/label/%s", f.user, category)},
	}

	_, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, addURL, []byte(formData.Encode()), f.headers, f.client,
	)
	if err != nil {
		return err
	}

	return nil
}

func (f FreshRSS) Enabled() bool {
	return f.isEnabled
}

func (f FreshRSS) RSSServerType() string {
	return f.rssType
}

// This function will authenticate to FreshRSS.
func authenticate(
	ctx context.Context,
	rssConfig config.RSSServerConfig,
	headers http.Header,
	logger *slog.Logger,
	client *http.Client,
) (string, error) {
	reqURL := fmt.Sprintf("%s/api/greader.php/accounts/ClientLogin", rssConfig.BaseURL)
	logger.Debug("Authenticating with FreshRSS", "url", reqURL)
	formData := []byte(
		url.Values{
			"Email":  {rssConfig.User},
			"Passwd": {rssConfig.Token},
		}.Encode(),
	)
	data, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, reqURL, formData, headers, client,
	)
	if err != nil {
		return "", err
	}

	var authToken string
	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		if after, ok := strings.CutPrefix(line, "Auth="); ok {
			authToken = after
		}
	}
	if authToken == "" {
		return "", fmt.Errorf("unable to parse FreshRSS auth response")
	}

	return authToken, nil
}

package rss

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/atomicmeganerd/starfeed/config"
)

// The freshRSS is a private struct that implements FreshRSSFeedManager.
type freshRSS struct {
	rssType   string
	baseURL   string
	user      string
	token     string
	enabled   bool
	client    *http.Client
	authToken string
}

// NewFreshRSSFeedManager creates a new FreshRSSFeedManager instance.
// Arguments:
// - cfg: This holds the configuration state this object needs.
// - client: The http client to use for requests (used for mocking).
func NewFreshRSSFeedManager(
	rssConfig *config.RSSServerConfig,
	client *http.Client,
) RSSServer {
	return &freshRSS{
		rssType: rssConfig.Type,
		baseURL: rssConfig.BaseURL,
		user:    rssConfig.User,
		token:   rssConfig.Token,
		enabled: rssConfig.Enabled,
		client:  client,
	}
}

func (f *freshRSS) Enabled() bool {
	return f.enabled
}

// Authenticate authenticates with the FreshRSS instance.
func (f *freshRSS) Authenticate(ctx context.Context) error {
	reqURL := fmt.Sprintf("%s/api/greader.php/accounts/ClientLogin", f.baseURL)
	slog.Debug("Authenticating with FreshRSS", "url", reqURL)

	formData := url.Values{
		"Email":  {f.user},
		"Passwd": {f.token},
	}

	data, err := f.doAPIRequest(ctx, reqURL, []byte(formData.Encode()), false)
	if err != nil {
		return err
	}

	authToken, err := parsePlainTextAuthResponse(data)
	if err != nil {
		return err
	}
	f.authToken = authToken

	slog.Debug("Authenticated with FreshRSS")
	return nil
}

// AddFeed adds a feed to the FreshRSS instance.
// Arguments:
// - feedURL: The URL of the feed to add.
// - name: The name of the feed.
// - category: The category to add the feed to.
func (f *freshRSS) AddFeed(
	ctx context.Context,
	feedURL, name, category string,
) error {
	addURL := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/quickadd", f.baseURL,
	)
	formData := url.Values{
		"quickadd": {feedURL},
	}

	slog.Debug("Adding feed to FreshRSS", "url", addURL)
	res, err := f.doAPIRequest(ctx, addURL, []byte(formData.Encode()), true)
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

	slog.Debug("Successfully added feed to FreshRSS", "feed", feedURL)
	return nil
}

// GetExistingFeeds gets the existing feeds from the FreshRSS instance.
func (f *freshRSS) GetExistingFeeds(ctx context.Context) (map[string]struct{}, error) {
	getURL := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/list?output=json", f.baseURL,
	)

	// Perform the request
	res, err := f.doAPIRequest(ctx, getURL, nil, true)
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
func (f *freshRSS) RemoveFeed(ctx context.Context, feedURL string) error {
	rmURL := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit", f.baseURL,
	)

	formData := url.Values{
		"ac": {"unsubscribe"},
		"s":  {fmt.Sprintf("feed/%s", feedURL)},
	}

	_, err := f.doAPIRequest(ctx, rmURL, []byte(formData.Encode()), true)
	if err != nil {
		return err
	}

	return nil
}

func (f *freshRSS) addFeedToCategory(
	ctx context.Context,
	streamId, name, category string,
) error {
	addURL := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit", f.baseURL,
	)

	formData := url.Values{
		"ac": {"edit"},
		"s":  {streamId},
		"t":  {name},
		"a":  {fmt.Sprintf("user/%s/label/%s", f.user, category)},
	}

	_, err := f.doAPIRequest(ctx, addURL, []byte(formData.Encode()), true)
	if err != nil {
		return err
	}

	return nil
}

func (f *freshRSS) doAPIRequest(
	ctx context.Context, url string, payload []byte, authHeader bool) ([]byte, error,
) {
	// Set headers
	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded",
	}
	if authHeader {
		headers["Authorization"] = fmt.Sprintf("GoogleLogin auth=%s", f.authToken)
	}

	// Create request (errors are ignored because the request is always valid)
	reader := bytes.NewReader(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, reader)
	if err != nil {
		return nil, err
	}

	// Process headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Make request
	res, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() // nolint: errcheck
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("FreshRSS returned an http error code %d", res.StatusCode)
	}

	return data, nil
}

func parsePlainTextAuthResponse(respData []byte) (string, error) {
	var authToken string
	lines := strings.SplitSeq(string(respData), "\n")
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

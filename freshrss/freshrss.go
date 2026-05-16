package freshrss

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

// FreshRSSFeedManager is an interface that manages the interaction with a FreshRSS instance.
type FreshRSSFeedManager interface {
	Authenticate(ctx context.Context) error
	AddFeed(ctx context.Context, feedUrl, name, category string) error
	GetExistingFeeds(ctx context.Context) (map[string]RSSFeed, error)
	RemoveFeed(ctx context.Context, feedUrl string) error
}

// The freshRSSFeedManager is a private struct that implements FreshRSSFeedManager.
type freshRSSFeedManager struct {
	cfg       *config.Config
	client    *http.Client
	authToken string
}

// NewFreshRSSFeedManager creates a new FreshRSSFeedManager instance.
// Arguments:
// - cfg: This holds the configuration state this object needs.
// - client: The http client to use for requests (used for mocking).
func NewFreshRSSFeedManager(
	cfg *config.Config,
	client *http.Client,
) FreshRSSFeedManager {
	return &freshRSSFeedManager{
		cfg:    cfg,
		client: client,
	}
}

// Authenticate authenticates with the FreshRSS instance.
func (f *freshRSSFeedManager) Authenticate(ctx context.Context) error {
	reqUrl := fmt.Sprintf("%s/api/greader.php/accounts/ClientLogin", f.cfg.FreshRSSURL)
	slog.Debug("Authenticating with FreshRSS", "url", reqUrl)

	formData := url.Values{
		"Email":  {f.cfg.FreshRSSUser},
		"Passwd": {f.cfg.FreshRSSToken},
	}

	data, err := f.doApiRequest(ctx, reqUrl, []byte(formData.Encode()), false)
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
// - feedUrl: The URL of the feed to add.
// - name: The name of the feed.
// - category: The category to add the feed to.
func (f *freshRSSFeedManager) AddFeed(
	ctx context.Context,
	feedUrl, name, category string,
) error {
	addUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/quickadd", f.cfg.FreshRSSURL,
	)
	formData := url.Values{
		"quickadd": {feedUrl},
	}

	slog.Debug("Adding feed to FreshRSS", "url", addUrl)
	res, err := f.doApiRequest(ctx, addUrl, []byte(formData.Encode()), true)
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

	slog.Debug("Successfully added feed to FreshRSS", "feed", feedUrl)
	return nil
}

// GetExistingFeeds gets the existing feeds from the FreshRSS instance.
func (f *freshRSSFeedManager) GetExistingFeeds(ctx context.Context) (map[string]RSSFeed, error) {
	getUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/list?output=json", f.cfg.FreshRSSURL,
	)

	// Perform the request
	res, err := f.doApiRequest(ctx, getUrl, nil, true)
	if err != nil {
		return nil, err
	}

	// Parse the response
	var feeds RSSFeedList
	if err = json.Unmarshal(res, &feeds); err != nil {
		return nil, err
	}

	feedMap := make(map[string]RSSFeed)
	for _, feed := range feeds.Feeds {
		feedMap[feed.Url] = feed
	}
	return feedMap, nil
}

// RemoveFeed removes a feed from the FreshRSS instance.
// Arguments:
// - feedUrl: The URL of the feed to remove.
func (f *freshRSSFeedManager) RemoveFeed(ctx context.Context, feedUrl string) error {
	rmUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit", f.cfg.FreshRSSURL,
	)

	formData := url.Values{
		"ac": {"unsubscribe"},
		"s":  {fmt.Sprintf("feed/%s", feedUrl)},
	}

	_, err := f.doApiRequest(ctx, rmUrl, []byte(formData.Encode()), true)
	if err != nil {
		return err
	}

	return nil
}

func (f *freshRSSFeedManager) addFeedToCategory(
	ctx context.Context,
	streamId, name, category string,
) error {
	addUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit", f.cfg.FreshRSSURL,
	)

	formData := url.Values{
		"ac": {"edit"},
		"s":  {streamId},
		"t":  {name},
		"a":  {fmt.Sprintf("user/%s/label/%s", f.cfg.FreshRSSUser, category)},
	}

	_, err := f.doApiRequest(ctx, addUrl, []byte(formData.Encode()), true)
	if err != nil {
		return err
	}

	return nil
}

func (f *freshRSSFeedManager) doApiRequest(
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

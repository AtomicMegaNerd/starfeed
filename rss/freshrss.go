package rss

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/atomicmeganerd/starfeed/common"
)

type FreshRSS struct {
	user    string
	url     string
	logger  *slog.Logger
	headers http.Header
	client  *http.Client
}

func NewFreshRSS(
	user, url string,
	logger *slog.Logger,
	client *http.Client,
) *FreshRSS {
	headers := http.Header{}
	headers.Set("Content-type", "application/x-www-form-urlencoded")
	return &FreshRSS{
		user:    user,
		url:     url,
		logger:  logger,
		headers: headers,
		client:  client,
	}
}

// This function will authenticate to FreshRSS.
func (f *FreshRSS) Authenticate(
	ctx context.Context,
	token string,
) error {
	reqURL := fmt.Sprintf("%s/api/greader.php/accounts/ClientLogin", f.url)
	f.logger.Debug("Authenticating to FreshRSS", "url", reqURL)
	formData := []byte(
		url.Values{
			"Email":  {f.user},
			"Passwd": {token},
		}.Encode(),
	)
	data, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, reqURL, formData, f.headers, f.client,
	)
	if err != nil {
		return fmt.Errorf("error authenticating to freshrss: %w, url: %s", err, reqURL)
	}

	var authToken string
	lines := strings.SplitSeq(string(data), "\n")
	for line := range lines {
		if after, ok := strings.CutPrefix(line, "Auth="); ok {
			authToken = after
		}
	}

	if authToken == "" {
		return errors.New("failed to parse authtoken returned from freshrss")
	}

	// We can set all required headers after we authenticate
	f.headers.Set("Authorization", fmt.Sprintf("GoogleLogin auth=%s", authToken))
	return nil
}

// Load all feeds that are under the given category.
func (f *FreshRSS) LoadFeeds(
	ctx context.Context, category FeedCategory,
) (RSSFeedSet, error) {
	newFeeds := make(RSSFeedSet)
	loadUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/list?output=json", f.url,
	)
	res, _, err := common.DoAPIRequest(ctx, http.MethodGet, loadUrl, nil, f.headers, f.client)
	if err != nil {
		return nil, err
	}

	// Parse the response
	feeds := &RSSFeedList{}
	if err = json.Unmarshal(res, &feeds); err != nil {
		return nil, err
	}

	for _, feed := range feeds.Feeds {
		// Only add feeds that are from the category that we care about
		for _, catStruct := range feed.Categories {
			if catStruct.Label == category {
				newFeeds[feed.URL] = struct{}{}
			}
		}
	}

	numFeeds := len(newFeeds)
	if numFeeds == 0 {
		f.logger.Warn("No feeds found in our RSS server", "numFeeds", numFeeds)
	} else {
		f.logger.Info(
			"Loaded existing feeds from FreshRSS", "numFeeds", numFeeds, "category", category,
		)
	}
	return newFeeds, nil
}

func (f *FreshRSS) AddFeed(
	ctx context.Context,
	feedURL common.FeedURL,
	name FeedName,
	category FeedCategory,
) error {

	addUrl := fmt.Sprintf("%s/api/greader.php/reader/api/0/subscription/quickadd", f.url)
	formData := url.Values{
		"quickadd": {feedURL.String()},
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
	if err = f.addFeedToCategory(ctx, name, category, feedResponse.StreamId); err != nil {
		return err
	}

	f.logger.Info("Successfully added feed", "feed", feedURL)
	return nil
}

func (f *FreshRSS) RemoveFeed(ctx context.Context, feedURL common.FeedURL) error {
	editUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit",
		f.url,
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

	f.logger.Info("Removed feed", "feed", feedURL)
	return nil
}

func (f *FreshRSS) addFeedToCategory(
	ctx context.Context,
	name FeedName,
	category FeedCategory,
	streamId string,
) error {
	addCategoryUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit",
		f.url,
	)
	formData := url.Values{
		"ac": {"edit"},
		"s":  {streamId},
		"t":  {name.String()},
		"a":  {fmt.Sprintf("user/%s/label/%s", f.user, category)},
	}

	if _, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, addCategoryUrl, []byte(formData.Encode()), f.headers, f.client,
	); err != nil {
		return err
	}
	return nil
}

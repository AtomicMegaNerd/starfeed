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
	"github.com/atomicmeganerd/starfeed/gitforge"
)

// RSS struct is for connecting to RSS servers. You can then Load/Add/Remove RSS
// feeds too/from the server.
type RSS struct {
	user    string
	url     string
	headers http.Header
	logger  *slog.Logger
	client  *http.Client
}

func NewRSS(
	user, url string,
	logger *slog.Logger,
	client *http.Client,
) *RSS {
	headers := http.Header{}
	headers.Set("Content-type", "application/x-www-form-urlencoded")
	return &RSS{
		user:    user,
		url:     url,
		headers: headers,
		logger:  logger,
		client:  client,
	}
}

// This function will authenticate to FreshRSS.
func (c *RSS) Authenticate(
	ctx context.Context,
	token string,
) error {
	reqURL := fmt.Sprintf("%s/api/greader.php/accounts/ClientLogin", c.url)
	formData := []byte(
		url.Values{
			"Email":  {c.user},
			"Passwd": {token},
		}.Encode(),
	)
	data, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, reqURL, formData, c.headers, c.client,
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
	c.headers.Set("Authorization", fmt.Sprintf("GoogleLogin auth=%s", authToken))
	return nil
}

func (c *RSS) load(
	ctx context.Context,
) (*FeedList, error) {
	loadUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/list?output=json", c.url,
	)
	res, _, err := common.DoAPIRequest(ctx, http.MethodGet, loadUrl, nil, c.headers, c.client)
	if err != nil {
		return nil, err
	}

	// Parse the response
	feedList := &FeedList{}
	if err = json.Unmarshal(res, &feedList); err != nil {
		return nil, err
	}

	return feedList, nil
}

func (c *RSS) unsubscribe(ctx context.Context, feed common.FeedURL) error {
	editUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit",
		c.url,
	)
	formData := url.Values{
		"ac": {"unsubscribe"},
		"s":  {fmt.Sprintf("feed/%s", feed)},
	}
	// We do not care about the response
	if _, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, editUrl, []byte(formData.Encode()), c.headers, c.client,
	); err != nil {
		return err
	}
	return nil
}

func (c *RSS) subscribe(
	ctx context.Context,
	req SubscribeRequest,
) error {
	addUrl := fmt.Sprintf("%s/api/greader.php/reader/api/0/subscription/quickadd", c.url)
	formData := url.Values{
		"quickadd": {req.URL.String()},
	}
	res, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, addUrl, []byte(formData.Encode()), c.headers, c.client,
	)
	if err != nil {
		return err
	}

	feedResponse := &SubscribeResponse{}
	if err = json.Unmarshal(res, &feedResponse); err != nil {
		return err
	}

	// Add the sub to the category
	if err = c.addFeedToCategory(ctx, req.Name, req.Category, feedResponse.StreamId); err != nil {
		return err
	}
	return nil
}

func (c *RSS) addFeedToCategory(
	ctx context.Context,
	name gitforge.RepoName,
	gitForgeName gitforge.ForgeName,
	streamId string,
) error {
	addCategoryUrl := fmt.Sprintf(
		"%s/api/greader.php/reader/api/0/subscription/edit",
		c.url,
	)
	formData := url.Values{
		"ac": {"edit"},
		"s":  {streamId},
		"t":  {name.String()},
		"a":  {fmt.Sprintf("user/%s/label/%s", c.user, gitForgeName)},
	}

	if _, _, err := common.DoAPIRequest(
		ctx, http.MethodPost, addCategoryUrl, []byte(formData.Encode()), c.headers, c.client,
	); err != nil {
		return err
	}
	return nil
}

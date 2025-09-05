package atom

import (
	"context"
	"encoding/xml"
	"io"
	"log/slog"
	"net/http"
)

// AtomFeedChecker is an interface for checking Atom feeds.
type AtomFeedChecker interface {
	CheckFeedHasEntries(feedUrl string) bool
}

// atomFeedChecker is the private implementation of AtomFeedChecker.
type atomFeedChecker struct {
	client *http.Client
	ctx    context.Context
}

func NewAtomFeedChecker(ctx context.Context, client *http.Client) AtomFeedChecker {
	return &atomFeedChecker{
		client: client,
		ctx:    ctx,
	}
}

// This function checks that the Atom feed has entries in it.
func (a *atomFeedChecker) CheckFeedHasEntries(feedUrl string) bool {
	// No request will always be valid here so we can ignore the error
	req, _ := http.NewRequestWithContext(a.ctx, "GET", feedUrl, nil)

	res, err := a.client.Do(req)
	if err != nil {
		slog.Error("Error making request to check ATOM feed", "feed", feedUrl, "error", err.Error())
		return false
	}
	defer res.Body.Close() // nolint:all

	data, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Error("Error reading response from ATOM feed", "feed", feedUrl, "error", err.Error())
		return false
	}

	var feed AtomFeed
	if err = xml.Unmarshal(data, &feed); err != nil {
		slog.Error("Error parsing XML data from response", "error", err.Error())
		return false
	}

	if len(feed.Entries) < 1 {
		return false
	}

	return true
}

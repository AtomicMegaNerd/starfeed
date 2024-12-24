package atom

import (
	"context"
	"encoding/xml"
	"io"
	"log/slog"
	"net/http"
)

type AtomFeedChecker struct {
	client *http.Client
	ctx    context.Context
}

func NewAtomFeedChecker(ctx context.Context, client *http.Client) *AtomFeedChecker {
	return &AtomFeedChecker{
		client: client,
		ctx:    ctx,
	}
}

// This function checks that the Atom feed has entries in it.
func (a *AtomFeedChecker) CheckFeedHasEntries(feedUrl string) bool {
	req, err := http.NewRequestWithContext(a.ctx, "GET", feedUrl, nil)
	if err != nil {
		slog.Error("Could not create request", "error", err.Error())
		return false
	}

	res, err := a.client.Do(req)
	if err != nil {
		slog.Error("Error making request to check ATOM feed", "feed", feedUrl, "error", err.Error())
		return false
	}
	defer res.Body.Close()

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

	if feed.Entries == nil || len(feed.Entries) < 1 {
		return false
	}

	return true
}

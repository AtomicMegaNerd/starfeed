package atom

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// AtomFeedChecker is an interface for checking Atom feeds.
type AtomFeedChecker interface {
	CheckFeedHasEntries(ctx context.Context, feedUrl string) (bool, error)
}

// atomFeedChecker is the private implementation of AtomFeedChecker.
type atomFeedChecker struct {
	client *http.Client
}

func NewAtomFeedChecker(client *http.Client) AtomFeedChecker {
	return &atomFeedChecker{
		client: client,
	}
}

// This function checks that the Atom feed has entries in it.
func (a *atomFeedChecker) CheckFeedHasEntries(
	ctx context.Context,
	feedUrl string,
) (bool, error) {
	// No request will always be valid here so we can ignore the error
	req, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)
	if err != nil {
		return false, err
	}

	res, err := a.client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close() // nolint:errcheck

	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("error response from ATOM feed %d", res.StatusCode)
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	var feed AtomFeed
	if err = xml.Unmarshal(data, &feed); err != nil {
		return false, err
	}

	if len(feed.Entries) < 1 {
		return false, nil
	}

	return true, nil
}

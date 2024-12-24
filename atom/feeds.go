package atom

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
)

type AtomFeedChecker struct {
	client *http.Client
	ctx    context.Context
}

func NewAtomFeedChecker(client *http.Client) *AtomFeedChecker {
	ctx := context.Background()
	return &AtomFeedChecker{
		client: client,
		ctx:    ctx,
	}
}

// This function checks that the Atom feed has entries in it.
func (a *AtomFeedChecker) CheckFeedHasEntries(feedUrl string) bool {
	req, err := http.NewRequestWithContext(a.ctx, "GET", feedUrl, nil)
	if err != nil {
		log.Errorf("Could not create request %s", err)
		return false
	}

	res, err := a.client.Do(req)
	if err != nil {
		log.Errorf("Error making request to check ATOM feed %s", err)
		return false
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Errorf("Error reading response from ATOM feed %s", err)
		return false
	}

	var feed AtomFeed
	err = xml.Unmarshal(data, &feed)
	if err != nil {
		log.Errorf("Error parsing XML data from response %s", err)
		return false
	}

	if feed.Entries == nil || len(feed.Entries) < 1 {
		return false
	}

	return true
}

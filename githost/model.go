package githost

import (
	"encoding/xml"
)

// This object represents a Git repo in a supported Git Host that is starred and that we want to
// get the Atom feed for.
type StarredRepo struct {
	Name    string `json:"name"`
	RepoURL string `json:"html_url"`
	FeedURL string `json:"feed_url"`
}

// This object represents an ATOM feed. We check to make sure that release feeds exist and
// do contain entries.
type AtomFeed struct {
	XMLName xml.Name `xml:"feed"`
	Title   string   `xml:"title"`
	Entries []Entry  `xml:"entry"`
}

type Entry struct {
	Title string `xml:"title"`
}

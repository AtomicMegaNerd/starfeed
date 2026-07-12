package gitforge

import (
	"encoding/xml"
)

const (
	GitHubForgeType  = "github"
	ForgejoForgeType = "forgejo"
)

// This object represents a Git repo in a supported Git Host that is starred and that we want to
// get the Atom feed for.
type GitRepo struct {
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

// This type both holds and validates the config for a GitForge
type GitForgeConfig struct {
	Type  string `validate:"required,oneof=github forgejo"`
	Name  string `validate:"required,min=3"`
	Fqdn  string `validate:"required,min=8"`
	Token string `validate:"required,min=10"`
}

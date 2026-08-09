package gitforge

import (
	"encoding/xml"

	"github.com/atomicmeganerd/starfeed/common"
)

const (
	GitHubForgeType  = "github"
	ForgejoForgeType = "forgejo"
)

type ForgeType string

func (t ForgeType) String() string {
	return string(t)
}

type ForgeName string

func (n ForgeName) String() string {
	return string(n)
}

type RepoName string

func (r RepoName) String() string {
	return string(r)
}

type RepoURL string

func (u RepoURL) String() string {
	return string(u)
}

type GitFeed struct {
	Name  RepoName
	URL   common.FeedURL
	Valid bool
}

// This object represents a Git repo in a supported Git Host that is starred and that we want to
// get the Atom feed for.
type Repo struct {
	Name RepoName `json:"name"`
	URL  RepoURL  `json:"html_url"`
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

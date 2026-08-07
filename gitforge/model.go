package gitforge

import (
	"encoding/xml"

	"github.com/atomicmeganerd/starfeed/common"
)

const (
	GitHubForgeType  = "github"
	ForgejoForgeType = "forgejo"
)

type GitRepoName string

func (r GitRepoName) String() string {
	return string(r)
}

type GitRepoURL string

func (u GitRepoURL) String() string {
	return string(u)
}

type StarredRepoMap map[common.FeedURL]GitRepoName

// This object represents a Git repo in a supported Git Host that is starred and that we want to
// get the Atom feed for.
type GitRepo struct {
	Name    GitRepoName    `json:"name"`
	RepoURL GitRepoURL     `json:"html_url"`
	FeedURL common.FeedURL `json:"feed_url"`
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

package gitforge

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"

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

type FeedResultMap map[common.FeedURL]GitRepoResult

// This object holds the result of our query. We need to distinguish between repos that have
// the following states:
//
//   - Querying the release feed returns 200 but there are no entries. We do not add such a feed
//     to RSS. Also if all entries are gone we should remove.
//   - Querying the release feed fails with 404. Not only do we not add such a feed but we will
//     remove such a feed from FreshRSS if the feed has been removed.
//   - Querying the reslease feed fails due to a GitHub issue (5xx) or network issue of some kind.
//     Here we need to make sure we don't remove from RSS because this could still be a valid feed
//     that is just being impacted by the outage.
type GitRepoResult struct {
	RepoName          GitRepoName
	RelFeedHasEntries bool
	Err               error
}

// Is stale means that querying the feed URL results in a 404 or the feed is there but has no
// entries. IsStale == true means this feed can be removed from FreshRSS.
func (r GitRepoResult) IsStale() bool {

	if r.Err == nil {
		// if we have no error but also no entries this repo is stale
		if !r.RelFeedHasEntries {
			return true
		}
		return false
	}

	// If we do have an error it is stale iff it is an HTTPError with code 404
	httpErr, ok := errors.AsType[common.HTTPError](r.Err)
	if !ok {
		return false
	}
	return httpErr.StatusCode == http.StatusNotFound
}

// Repos are ok if there was no error querying the feed and the feed has entries
func (r GitRepoResult) IsOK() bool {
	return r.Err == nil && r.RelFeedHasEntries
}

// Equal compares two GitRepoResult values. Errors are considered equal if they have the same type.
func (r GitRepoResult) Equal(other GitRepoResult) bool {
	if r.RepoName != other.RepoName {
		return false
	}
	if r.RelFeedHasEntries != other.RelFeedHasEntries {
		return false
	}

	if r.Err == nil && other.Err == nil {
		return true
	}
	if r.Err == nil || other.Err == nil {
		return false
	}

	return fmt.Sprintf("%T", r.Err) == fmt.Sprintf("%T", other.Err)
}

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

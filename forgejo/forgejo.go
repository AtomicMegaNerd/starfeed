package forgejo

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/atomicmeganerd/starfeed/githost"
)

type forgejoStarredFeedBuilder struct {
	gitHost           githost.GitHostConfig
	client            *http.Client
	nextPageLinkRegex *regexp.Regexp
	isRelRepoRegex    *regexp.Regexp
}

func NewForgejoStarredFeedBuilder(
	gitHost githost.GitHostConfig,
	client *http.Client,
) githost.GitHost {
	// This regex is used to find the next page link in the Forgejo API response
	nextPageLinkRegex := regexp.MustCompile(`<([^>]+)>; rel="next"`)
	// This regex is used to determine if an RSS feed is a Forgejo release feed
	isRelRepoRegex := regexp.MustCompile(
		fmt.Sprintf(`^%s/[\w\.\-]+/[\w\.\-]+/releases\.atom`, regexp.QuoteMeta(gitHost.BaseURL)),
	)

	return &forgejoStarredFeedBuilder{gitHost, client, nextPageLinkRegex, isRelRepoRegex}
}

func (f *forgejoStarredFeedBuilder) GetStarredRepos(
	ctx context.Context,
) (map[string]githost.Repo, error) {
	return nil, nil
}

func (f *forgejoStarredFeedBuilder) IsReleaseFeed(feedUrl string) bool {
	return false
}

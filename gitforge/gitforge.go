package gitforge

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/atomicmeganerd/starfeed/common"
	"golang.org/x/sync/errgroup"
)

// This regex will match if there is a next page in the response headers
var nextPagePattern = regexp.MustCompile(`<([^>]+)>; rel="next"`)

type GitForge struct {
	fetchRepoURL string
	headers      http.Header
	logger       *slog.Logger
	client       *http.Client
}

func NewGitForge(
	forgeType, fqdn, token string,
	logger *slog.Logger,
	client *http.Client,
) GitForge {
	return GitForge{
		fetchRepoURL: buildStarredRepoUrl(forgeType, fqdn),
		headers:      buildHeaders(forgeType, token),
		logger:       logger,
		client:       client,
	}
}

func (g GitForge) LoadFeeds(
	ctx context.Context,
) (StarredRepoMap, error) {
	// Get all starredRepos
	starredRepos, err := g.fetchStarredRepos(ctx)
	if err != nil {
		return nil, err
	}

	// We aren't using errors here but errgroup gives us SetLimit
	eg := &errgroup.Group{}
	eg.SetLimit(5)

	// This is the list of starredFeeds that we will return to the caller
	starredFeeds := make(StarredRepoMap)

	// Repos that have release feeds will be sent to this channel. Repos sent to this channel will
	// then be added to the feeds map. There is no need for a buffered channel here as the
	// consumer basically does nothing except writing to a map.
	repoChan := make(chan GitRepo)

	var wg sync.WaitGroup
	wg.Go(func() {
		// This for loop consumes each message that is received. It blocks if the channel is open
		// but is waiting for a message. When the channel is closed the range is complete and the
		// for loop terminates.
		for repo := range repoChan {
			starredFeeds[repo.FeedURL] = repo.Name
			g.logger.Debug("Added feed to map", "repo", repo.Name, "feed", repo.FeedURL)
		}
	})

	// Check each repo to make sure it has valid entries in its ATOM feed for releases
	// This can be done in parallel to make it much faster. Send each release repo to the channel
	for _, repo := range starredRepos {
		eg.Go(func() error {
			logger := g.logger.With(
				"repo", repo.Name,
				"feed", repo.FeedURL,
			)

			if g.repoHasReleaseFeed(ctx, repo) {
				logger.Debug("Trying to send repo to channel")
				repoChan <- repo
				logger.Info("Adding feed for repo to feeds map")
				return nil
			}

			logger.Warn("Repo does not have valid release feed")
			return nil
		})
	}

	// When the producers are done, close the channel
	_ = eg.Wait()
	close(repoChan)

	// Wait for the consumer to receive all messages from the producers
	wg.Wait()

	g.logger.Info("Successfully added all feeds to feeds map", "numFeeds", len(starredFeeds))
	return starredFeeds, nil
}

func (g GitForge) fetchStarredRepos(
	ctx context.Context,
) ([]GitRepo, error) {
	allRepos := make([]GitRepo, 0)
	nextPageURL := g.fetchRepoURL
	for {
		g.logger.Debug("Fetching starred repos", "url", nextPageURL)
		data, respHeaders, err := common.DoAPIRequest(
			ctx,
			http.MethodGet,
			nextPageURL,
			nil,
			g.headers,
			g.client,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"error %w getting raw data from gitforge url: %s", err, nextPageURL,
			)
		}

		repos := make([]GitRepo, 0)
		if err := json.Unmarshal(data, &repos); err != nil {
			return nil, fmt.Errorf(
				"error %w parsing JSON response from gitforge", err,
			)
		}

		for ix := range repos {
			repos[ix].FeedURL = common.FeedURL(
				fmt.Sprintf(
					"%s/releases.atom", repos[ix].RepoURL,
				),
			)
		}
		allRepos = append(allRepos, repos...)

		nextPageURL = g.parseNextPageURL(respHeaders)
		if nextPageURL == "" {
			g.logger.Info("Finished loading starred repos", "numRepos", len(allRepos))
			return allRepos, nil
		}
	}
}

func (g GitForge) repoHasReleaseFeed(
	ctx context.Context,
	repo GitRepo,
) bool {
	logger := g.logger.With("repo", repo.Name, "feed", repo.FeedURL)
	logger.Debug("Checking if repo has release feed")
	data, _, err := common.DoAPIRequest(
		ctx,
		http.MethodGet,
		repo.FeedURL.String(),
		nil,
		g.headers,
		g.client,
	)
	if err != nil {
		return false
	}
	relFeed := &AtomFeed{}
	if err = xml.Unmarshal(data, relFeed); err != nil {
		return false
	}
	if len(relFeed.Entries) >= 1 {
		logger.Debug("Repo feed is valid")
		return true
	}
	return false
}

func (g GitForge) parseNextPageURL(respHeaders http.Header) string {
	linkHeader := respHeaders.Get("Link")
	if linkHeader == "" {
		return ""
	}

	g.logger.Debug("linkHeader found", "linkHeader", linkHeader)
	links := strings.SplitSeq(linkHeader, ",")
	for link := range links {
		matches := nextPagePattern.FindStringSubmatch(link)
		if len(matches) == 2 {
			return matches[1]
		}
	}
	return ""
}

func buildHeaders(forgeType, token string) http.Header {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept", "application/json")
	headers.Set("User-Agent", "github.com/atomicmeganerd/starfeed")
	headers.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if forgeType == GitHubForgeType {
		headers.Set("X-GitHub-Api-Version", "2022-11-28")
	}
	return headers
}

func buildStarredRepoUrl(forgeType, fqdn string) string {
	if forgeType == GitHubForgeType {
		return fmt.Sprintf("https://api.%s/user/starred?per_page=100", fqdn)
	}
	return fmt.Sprintf("https://%s/api/v1/user/starred?limit=100", fqdn)
}

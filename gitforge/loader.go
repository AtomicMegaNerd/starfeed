package gitforge

import (
	"context"
	"log/slog"
)

type GitForgeLoader struct {
	Name         GitForgeName
	LoadChan     chan struct{}
	GitFeedsChan chan GitFeed
	forge        GitForge
	stop         context.CancelFunc
	logger       *slog.Logger
}

func NewGitForgeLoader(
	forge GitForge,
	stop context.CancelFunc,
	logger *slog.Logger,
) GitForgeLoader {

	loadChan := make(chan struct{}, 1)
	gitFeedsChan := make(chan GitFeed)

	return GitForgeLoader{
		Name:         forge.Name,
		LoadChan:     loadChan,
		GitFeedsChan: gitFeedsChan,
		forge:        forge,
		stop:         stop,
		logger:       logger,
	}
}

// This method registers our GitForge so it listens for messages
func (c GitForgeLoader) Init(ctx context.Context) {
	for {
		select {
		case <-c.LoadChan:
			repos, err := c.forge.getRepos(ctx)
			if err != nil {
				c.stop()
				return
			}
			sem := make(chan struct{}, 5)
			for _, repo := range repos {
				go func() {
					sem <- struct{}{}
					c.GitFeedsChan <- c.forge.buildFeedFromRepo(ctx, repo)
				}()
			}
		case <-ctx.Done():
			return
		}
	}
}

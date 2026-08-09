package gitforge

import (
	"context"
	"log/slog"
)

type Loader struct {
	Name     ForgeName
	LoadChan chan struct{}
	FeedChan chan GitFeed
	forge    GitForge
	stop     context.CancelFunc
	logger   *slog.Logger
}

func NewLoader(
	forge GitForge,
	stop context.CancelFunc,
	logger *slog.Logger,
) Loader {
	return Loader{
		Name:     forge.Name,
		LoadChan: make(chan struct{}, 1),
		FeedChan: make(chan GitFeed, 5),
		forge:    forge,
		stop:     stop,
		logger:   logger,
	}
}

// This method registers our GitForge so it listens for messages
func (c Loader) Init(ctx context.Context) {
	for {
		select {
		case <-c.LoadChan:
			repos, err := c.forge.load(ctx)
			if err != nil {
				c.stop()
				return
			}
			sem := make(chan struct{}, 5)
			for _, repo := range repos {
				go func() {
					sem <- struct{}{}
					c.FeedChan <- c.forge.rssFeedFromRepo(ctx, repo)
				}()
			}
		case <-ctx.Done():
			return
		}
	}
}

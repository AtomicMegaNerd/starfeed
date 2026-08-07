package runners

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Defining an interface here for any Runner that we want to add to our runnerSlice
type StarfeedRunner interface {
	Run(ctx context.Context) error
}

// Here we execute the runners in parallel...
func ExecuteRunners(ctx context.Context, runners []StarfeedRunner) error {
	errGroup, runnerCtx := errgroup.WithContext(ctx)
	for _, runner := range runners {
		errGroup.Go(func() error {
			return runner.Run(runnerCtx)
		})
	}
	return errGroup.Wait()
}

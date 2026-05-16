package runner

import "context"

// The Runner interface is for each major workflow in our app which will be excuted.
type Runner interface {
	Run(ctx context.Context) error
}

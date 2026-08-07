package runners

import (
	"context"
	"errors"
	"testing"
	"time"
)

// timeout is how long before the test cancels the context for cases
// that exercise external cancellation.
const timeout = 50 * time.Millisecond

// mockRunner is a minimal StarfeedRunner for testing ExecuteRunners. It can
// optionally return an error or block until the context is cancelled.
type mockRunner struct {
	err              error
	sleep            bool
	blockUntilCancel bool
}

func (m mockRunner) Run(ctx context.Context) error {
	// If we enable sleep we want to sleep past our timeout before we block
	// this will trigger the deferred cancel() call below.
	if m.sleep {
		time.Sleep(timeout + 10*time.Millisecond)
	}
	if m.blockUntilCancel {
		<-ctx.Done()
		return ctx.Err()
	}
	return m.err
}

func TestExecuteRunners(t *testing.T) {
	t.Parallel()

	mockErr := errors.New("runner failed")

	defaultCtxFunc := func() (context.Context, context.CancelFunc) {
		return context.Background(), nil
	}

	testCases := []struct {
		name    string
		runners []StarfeedRunner
		// We have to return a function here so that we do not start the timeout when defining
		// the test cases. Instead we'll set it when the test executes.
		ctxFunc   func() (context.Context, context.CancelFunc)
		expectErr error
	}{
		{
			name:      "empty slice returns nil",
			runners:   nil,
			ctxFunc:   defaultCtxFunc,
			expectErr: nil,
		},
		{
			name: "all runners succeed",
			runners: []StarfeedRunner{
				&mockRunner{},
				&mockRunner{},
				&mockRunner{},
			},
			ctxFunc:   defaultCtxFunc,
			expectErr: nil,
		},
		{
			name: "one runner fails",
			runners: []StarfeedRunner{
				&mockRunner{},
				&mockRunner{err: mockErr},
				&mockRunner{},
			},
			ctxFunc:   defaultCtxFunc,
			expectErr: mockErr,
		},
		{
			name: "blocking runner exits on sibling error",
			runners: []StarfeedRunner{
				&mockRunner{blockUntilCancel: true},
				&mockRunner{err: mockErr},
			},
			ctxFunc:   defaultCtxFunc,
			expectErr: mockErr,
		},
		{
			name: "blocking runner exits on timeout exceeded",
			runners: []StarfeedRunner{
				&mockRunner{blockUntilCancel: true, sleep: true},
			},
			ctxFunc: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), timeout)
			},
			expectErr: context.DeadlineExceeded,
		},
		{
			name: "blocking runner exits on cancel called",
			runners: []StarfeedRunner{
				&mockRunner{blockUntilCancel: true},
			},
			ctxFunc: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			expectErr: context.Canceled,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := tc.ctxFunc()
			// In the case of a timeout we will sleep past the deadline and cancel will get
			// triggered. Having a deadline value means that a timeout was set.
			if cancel != nil {
				_, timeoutWasSet := ctx.Deadline()

				if timeoutWasSet {
					// NOTE: This defer cancel() call is not needed for DeadlineExceeded to be sent
					// to the context, it is just good hygeine for cleanup.
					defer cancel()
				} else {
					go func() {
						// This gouroutine will call cancel() after a short delay so that the
						// runners will all start and be parked before cancel() is triggered.
						time.Sleep(20 * time.Millisecond)
						cancel()
					}()
				}

			}

			err := ExecuteRunners(ctx, tc.runners)

			if tc.expectErr == nil && err != nil {
				t.Fatalf("Unexpected error %q", err)
			}
			if tc.expectErr != nil && !errors.Is(err, tc.expectErr) {
				t.Fatalf("Expected %v but got %v", tc.expectErr, err)
			}
		})
	}
}

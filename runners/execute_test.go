package runners

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// mockRunner is a minimal StarfeedRunner for testing ExecuteRunners. It records
// how many times Run was called and can optionally return an error or block
// until the context is cancelled.
type mockRunner struct {
	calls            atomic.Int32
	err              error
	blockUntilCancel bool
}

func (m *mockRunner) Run(ctx context.Context) error {
	m.calls.Add(1)
	if m.blockUntilCancel {
		<-ctx.Done()
		return ctx.Err()
	}
	return m.err
}

func TestExecuteRunners(t *testing.T) {
	t.Parallel()

	errBoom := errors.New("boom")

	testCases := []struct {
		name        string
		runners     []*mockRunner
		expectError bool
		expectCalls int32
	}{
		{
			name:        "empty slice returns nil",
			runners:     nil,
			expectError: false,
			expectCalls: 0,
		},
		{
			name: "all runners succeed",
			runners: []*mockRunner{
				{},
				{},
				{},
			},
			expectError: false,
			expectCalls: 3,
		},
		{
			name: "one runner fails",
			runners: []*mockRunner{
				{},
				{err: errBoom},
				{},
			},
			expectError: true,
		},
		{
			name: "runners respect ctx cancellation on error",
			runners: []*mockRunner{
				{blockUntilCancel: true},
				{err: errBoom},
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			starfeedRunners := make([]StarfeedRunner, len(tc.runners))
			for i, r := range tc.runners {
				starfeedRunners[i] = r
			}

			err := ExecuteRunners(ctx, starfeedRunners)

			if tc.expectError && err == nil {
				t.Fatalf("Expected error but got none")
			}
			if !tc.expectError && err != nil {
				t.Fatalf("Unexpected error %q", err)
			}

			if tc.expectCalls > 0 {
				var totalCalls int32
				for _, r := range tc.runners {
					totalCalls += r.calls.Load()
				}
				if totalCalls != tc.expectCalls {
					t.Fatalf("Expected %d total calls but got %d", tc.expectCalls, totalCalls)
				}
			}
		})
	}
}

// TestExecuteRunnersCancellation verifies that a blocking runner exits promptly
// when the context is cancelled externally (not via a sibling error).
func TestExecuteRunnersCancellation(t *testing.T) {
	t.Parallel()

	blocker := &mockRunner{blockUntilCancel: true}
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := ExecuteRunners(ctx, []StarfeedRunner{blocker})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Expected context.Canceled but got %v", err)
	}
	if blocker.calls.Load() != 1 {
		t.Fatalf("Expected blocker to have been called once, got %d", blocker.calls.Load())
	}
}

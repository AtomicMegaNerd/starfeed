package gitforge

import (
	"errors"
	"net/http"
	"testing"

	"github.com/atomicmeganerd/starfeed/common"
)

func TestGitRepoResultIsStale(t *testing.T) {
	testCases := []struct {
		name     string
		result   GitRepoResult
		expected bool
	}{
		{
			name:     "No error and no entries is stale",
			result:   GitRepoResult{RepoName: "repo1"},
			expected: true,
		},
		{
			name:     "No error and has entries is not stale",
			result:   GitRepoResult{RepoName: "repo1", RelFeedHasEntries: true},
			expected: false,
		},
		{
			name: "HTTP 404 error is stale",
			result: GitRepoResult{
				RepoName: "repo1",
				Err:      common.HTTPError{StatusCode: http.StatusNotFound},
			},
			expected: true,
		},
		{
			name: "HTTP 500 error is not stale",
			result: GitRepoResult{
				RepoName: "repo1",
				Err:      common.HTTPError{StatusCode: http.StatusInternalServerError},
			},
			expected: false,
		},
		{
			name:     "Non-HTTP error is not stale",
			result:   GitRepoResult{RepoName: "repo1", Err: errors.New("network failure")},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.result.IsStale(); got != tc.expected {
				t.Errorf("IsStale() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestGitRepoResultIsOK(t *testing.T) {
	testCases := []struct {
		name     string
		result   GitRepoResult
		expected bool
	}{
		{
			name:     "No error and has entries is ok",
			result:   GitRepoResult{RepoName: "repo1", RelFeedHasEntries: true},
			expected: true,
		},
		{
			name:     "No error and no entries is not ok",
			result:   GitRepoResult{RepoName: "repo1"},
			expected: false,
		},
		{
			name: "Error with entries is not ok",
			result: GitRepoResult{
				RepoName:          "repo1",
				RelFeedHasEntries: true,
				Err:               common.HTTPError{StatusCode: http.StatusNotFound},
			},
			expected: false,
		},
		{
			name: "Error without entries is not ok",
			result: GitRepoResult{
				RepoName: "repo1",
				Err:      common.HTTPError{StatusCode: http.StatusNotFound},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.result.IsOK(); got != tc.expected {
				t.Errorf("IsOK() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestGitRepoResultEqual(t *testing.T) {
	testCases := []struct {
		name     string
		a        GitRepoResult
		b        GitRepoResult
		expected bool
	}{
		{
			name:     "Both zero values are equal",
			a:        GitRepoResult{},
			b:        GitRepoResult{},
			expected: true,
		},
		{
			name:     "Same fields are equal",
			a:        GitRepoResult{RepoName: "repo1", RelFeedHasEntries: true},
			b:        GitRepoResult{RepoName: "repo1", RelFeedHasEntries: true},
			expected: true,
		},
		{
			name:     "Different repo name is not equal",
			a:        GitRepoResult{RepoName: "repo1"},
			b:        GitRepoResult{RepoName: "repo2"},
			expected: false,
		},
		{
			name:     "Different entries flag is not equal",
			a:        GitRepoResult{RepoName: "repo1", RelFeedHasEntries: true},
			b:        GitRepoResult{RepoName: "repo1"},
			expected: false,
		},
		{
			name: "Same error type is equal",
			a: GitRepoResult{
				RepoName: "repo1",
				Err:      common.HTTPError{StatusCode: http.StatusNotFound},
			},
			b: GitRepoResult{
				RepoName: "repo1",
				Err:      common.HTTPError{StatusCode: http.StatusInternalServerError},
			},
			expected: true,
		},
		{
			name: "Different error types are not equal",
			a: GitRepoResult{
				RepoName: "repo1",
				Err:      common.HTTPError{StatusCode: http.StatusNotFound},
			},
			b: GitRepoResult{
				RepoName: "repo1",
				Err:      errors.New("network failure"),
			},
			expected: false,
		},
		{
			name: "One nil error and one non-nil error is not equal",
			a:    GitRepoResult{RepoName: "repo1"},
			b: GitRepoResult{
				RepoName: "repo1",
				Err:      common.HTTPError{StatusCode: http.StatusNotFound},
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := tc.a.Equal(tc.b); got != tc.expected {
				t.Errorf("Equal() = %v, want %v", got, tc.expected)
			}
		})
	}
}

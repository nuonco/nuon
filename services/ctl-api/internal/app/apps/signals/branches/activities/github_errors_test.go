package activities

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

func ghErrorResponse(status int, message string) error {
	return &github.ErrorResponse{
		Response: &http.Response{StatusCode: status},
		Message:  message,
	}
}

func TestNonRetryableGitHubError(t *testing.T) {
	for _, tc := range []struct {
		name          string
		err           error
		wantNonRetry  bool
		wantErrorType string
	}{
		{
			// GitHub answers an unknown ref on the commits endpoint with 422.
			name:          "422 no commit found for ref",
			err:           ghErrorResponse(http.StatusUnprocessableEntity, "No commit found for SHA: nope"),
			wantNonRetry:  true,
			wantErrorType: "github_422",
		},
		{
			name:          "404 repo gone",
			err:           ghErrorResponse(http.StatusNotFound, "Not Found"),
			wantNonRetry:  true,
			wantErrorType: "github_404",
		},
		{
			name:          "401 revoked token",
			err:           ghErrorResponse(http.StatusUnauthorized, "Bad credentials"),
			wantNonRetry:  true,
			wantErrorType: "github_401",
		},
		{
			name:         "500 stays retryable",
			err:          ghErrorResponse(http.StatusInternalServerError, "Server Error"),
			wantNonRetry: false,
		},
		{
			name:         "rate limit stays retryable",
			err:          &github.RateLimitError{Response: &http.Response{StatusCode: http.StatusForbidden}},
			wantNonRetry: false,
		},
		{
			name:         "secondary rate limit stays retryable",
			err:          &github.AbuseRateLimitError{Response: &http.Response{StatusCode: http.StatusForbidden}},
			wantNonRetry: false,
		},
		{
			name:         "non-github error",
			err:          fmt.Errorf("dial tcp: connection refused"),
			wantNonRetry: false,
		},
		{
			name:         "error response without a response is not classified",
			err:          &github.ErrorResponse{Message: "hand rolled"},
			wantNonRetry: false,
		},
		{
			// fetchLatestCommit reaches the API through helpers that wrap in
			// stderr.ErrUser and fmt.Errorf.
			name:          "4xx found through wrapping",
			err:           stderr.ErrUser{Err: fmt.Errorf("unable to get latest commit: %w", ghErrorResponse(http.StatusUnprocessableEntity, "No commit found"))},
			wantNonRetry:  true,
			wantErrorType: "github_422",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := nonRetryableGitHubError(tc.err)

			if !tc.wantNonRetry {
				assert.Nil(t, got)
				return
			}

			require.NotNil(t, got)

			var appErr *temporal.ApplicationError
			require.ErrorAs(t, got, &appErr)
			assert.True(t, appErr.NonRetryable())
			assert.Equal(t, tc.wantErrorType, appErr.Type())
		})
	}
}

// Temporal's DefaultFailureConverter type-switches on the concrete error rather
// than using errors.As, so wrapping a non-retryable error makes the top-level
// failure retryable again. This guards the unwrapped return in
// FetchInstallSyncCommit.
func TestWrappingDefeatsNonRetryable(t *testing.T) {
	nonRetryable := nonRetryableGitHubError(ghErrorResponse(http.StatusUnprocessableEntity, "No commit found"))
	require.NotNil(t, nonRetryable)

	var direct *temporal.ApplicationError
	require.True(t, assertConcreteApplicationError(nonRetryable, &direct))
	assert.True(t, direct.NonRetryable())

	var wrapped *temporal.ApplicationError
	assert.False(t,
		assertConcreteApplicationError(fmt.Errorf("unable to fetch commit: %w", nonRetryable), &wrapped),
		"wrapping must not be reintroduced in activities that return this error",
	)
}

// assertConcreteApplicationError mirrors the SDK's type switch: a match only
// happens when the error itself is an *ApplicationError, not when one is
// somewhere down the unwrap chain.
func assertConcreteApplicationError(err error, target **temporal.ApplicationError) bool {
	appErr, ok := err.(*temporal.ApplicationError)
	if ok {
		*target = appErr
	}
	return ok
}

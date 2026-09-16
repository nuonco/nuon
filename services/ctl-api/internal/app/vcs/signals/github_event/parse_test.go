package githubevent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePullRequestEventDraftAndHeadRef(t *testing.T) {
	info, err := parsePullRequestEvent(map[string]any{
		"action": "opened",
		"pull_request": map[string]any{
			"number": float64(13),
			"draft":  true,
			"base":   map[string]any{"ref": "main"},
			"head": map[string]any{
				"sha": "abc123",
				"ref": "jm/test-ci",
			},
		},
		"repository": map[string]any{"full_name": "acme/app"},
	})
	require.NoError(t, err)
	require.Equal(t, "main", info.BaseBranch)
	require.Equal(t, "abc123", info.HeadSHA)
	require.Equal(t, "jm/test-ci", info.HeadRef)
	require.True(t, info.Draft)
	require.Equal(t, "opened", info.Action)
}

func TestParsePullRequestEventReadyForReview(t *testing.T) {
	info, err := parsePullRequestEvent(map[string]any{
		"action": "ready_for_review",
		"pull_request": map[string]any{
			"number": float64(13),
			"draft":  false,
			"base":   map[string]any{"ref": "main"},
			"head": map[string]any{
				"sha": "def456",
				"ref": "jm/test-ci",
			},
		},
		"repository": map[string]any{"full_name": "acme/app"},
	})
	require.NoError(t, err)
	require.False(t, info.Draft)
	require.Equal(t, "ready_for_review", info.Action)
}

func TestParseTagPush(t *testing.T) {
	info, err := parsePushEvent(map[string]any{
		"ref":        "refs/tags/foobar/v0.0.1",
		"after":      "abc123",
		"repository": map[string]any{"full_name": "acme/app"},
	})
	require.NoError(t, err)
	require.Empty(t, info.Branch)
	require.Equal(t, "foobar/v0.0.1", info.Tag)
	require.Equal(t, "abc123", info.HeadSHA)
	require.False(t, info.Deleted)
}

func TestParseDeletedTagPush(t *testing.T) {
	info, err := parsePushEvent(map[string]any{
		"ref":        "refs/tags/foobar/v0.0.1",
		"after":      "0000000000000000000000000000000000000000",
		"deleted":    true,
		"repository": map[string]any{"full_name": "acme/app"},
	})
	require.NoError(t, err)
	require.True(t, info.Deleted)
}

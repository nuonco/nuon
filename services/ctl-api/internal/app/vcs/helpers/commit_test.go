package helpers

import (
	"testing"
	"time"

	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/require"
)

func TestGithubCommitToVCSConnectionCommitUsesCommitterTime(t *testing.T) {
	committed := time.Date(2026, 10, 1, 15, 4, 0, 0, time.UTC)
	authored := committed.Add(-48 * time.Hour)
	sha := "abc123"
	name := "Casey"
	email := "casey@example.com"
	message := "feat: add cache"
	avatar := "https://example.com/a.png"

	got := (*Helpers)(nil).GithubCommitToVCSConnectionCommit(&github.RepositoryCommit{
		SHA: &sha,
		Commit: &github.Commit{
			Message: &message,
			Author: &github.CommitAuthor{
				Name:  &name,
				Email: &email,
				Date:  &github.Timestamp{Time: authored},
			},
			Committer: &github.CommitAuthor{
				Date: &github.Timestamp{Time: committed},
			},
		},
		Author: &github.User{AvatarURL: &avatar},
	}, "owner", "ConnectedGithubVCSConfig", "vcs-1")

	require.NotNil(t, got)
	require.NotNil(t, got.CommittedAt)
	require.True(t, got.CommittedAt.Equal(committed))
	require.Equal(t, name, got.AuthorName)
	require.Equal(t, email, got.AuthorEmail)
	require.Equal(t, message, got.Message)
	require.Equal(t, avatar, got.AuthorAvatarURL)
}

func TestGithubCommitToVCSConnectionCommitFallsBackToAuthorTime(t *testing.T) {
	authored := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	sha := "def456"

	got := (*Helpers)(nil).GithubCommitToVCSConnectionCommit(&github.RepositoryCommit{
		SHA: &sha,
		Commit: &github.Commit{
			Author: &github.CommitAuthor{Date: &github.Timestamp{Time: authored}},
		},
	}, "owner", "PublicGitVCSConfig", "")

	require.NotNil(t, got)
	require.NotNil(t, got.CommittedAt)
	require.True(t, got.CommittedAt.Equal(authored))
}

func TestGithubCommitToVCSConnectionCommitOmitsMissingTime(t *testing.T) {
	sha := "ghi789"

	got := (*Helpers)(nil).GithubCommitToVCSConnectionCommit(&github.RepositoryCommit{
		SHA:    &sha,
		Commit: &github.Commit{},
	}, "owner", "PublicGitVCSConfig", "")

	require.NotNil(t, got)
	require.Nil(t, got.CommittedAt)
}

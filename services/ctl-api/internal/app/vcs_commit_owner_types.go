package app

// VCSCommitOwnerType represents the type of entity that owns a VCS commit.
// VCS commits use polymorphic ownership to reference either ConnectedGithubVCSConfig
// or PublicGitVCSConfig entities.
type VCSCommitOwnerType string

const (
	// VCSCommitOwnerTypeConnectedGithubVCSConfig indicates the commit is owned by a ConnectedGithubVCSConfig
	// (authenticated GitHub repository via GitHub App installation)
	VCSCommitOwnerTypeConnectedGithubVCSConfig VCSCommitOwnerType = "connected_github_vcs_configs"

	// VCSCommitOwnerTypePublicGitVCSConfig indicates the commit is owned by a PublicGitVCSConfig
	// (public GitHub repository with no authentication)
	VCSCommitOwnerTypePublicGitVCSConfig VCSCommitOwnerType = "public_git_vcs_configs"
)

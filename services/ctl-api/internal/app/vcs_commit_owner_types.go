package app

type VCSCommitOwnerType string

const (
	VCSCommitOwnerTypeConnectedGithubVCSConfig VCSCommitOwnerType = "connected_github_vcs_configs"

	VCSCommitOwnerTypePublicGitVCSConfig VCSCommitOwnerType = "public_git_vcs_configs"
)

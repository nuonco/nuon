package activities

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func connectedRun(repo, branch string) *app.AppBranchRun {
	return &app.AppBranchRun{
		AppBranchConfig: app.AppBranchConfig{
			ConnectedGithubVCSConfig: &app.ConnectedGithubVCSConfig{Repo: repo, Branch: branch},
		},
	}
}

func connectedBuild(repo, branch string) *app.ComponentBuild {
	return &app.ComponentBuild{
		ComponentConfigConnection: app.ComponentConfigConnection{
			ConnectedGithubVCSConfig: &app.ConnectedGithubVCSConfig{Repo: repo, Branch: branch},
		},
	}
}

func publicBuild(repo, branch string, gitRef *string) *app.ComponentBuild {
	return &app.ComponentBuild{
		GitRef: gitRef,
		ComponentConfigConnection: app.ComponentConfigConnection{
			PublicGitVCSConfig: &app.PublicGitVCSConfig{Repo: repo, Branch: branch},
		},
	}
}

func TestBuildTracksBranchSource(t *testing.T) {
	t.Run("same connected repo and branch", func(t *testing.T) {
		require.True(t, buildTracksBranchSource(
			connectedRun("acme/infra", "main"),
			connectedBuild("acme/infra", "main"),
		))
	})

	t.Run("repo casing is ignored", func(t *testing.T) {
		require.True(t, buildTracksBranchSource(
			connectedRun("Acme/Infra", "main"),
			connectedBuild("acme/infra", "main"),
		))
	})

	t.Run("different repo", func(t *testing.T) {
		require.False(t, buildTracksBranchSource(
			connectedRun("acme/infra", "main"),
			connectedBuild("acme/services", "main"),
		))
	})

	t.Run("different branch", func(t *testing.T) {
		require.False(t, buildTracksBranchSource(
			connectedRun("acme/infra", "main"),
			connectedBuild("acme/infra", "release"),
		))
	})

	t.Run("same public repo and branch", func(t *testing.T) {
		run := &app.AppBranchRun{
			AppBranchConfig: app.AppBranchConfig{
				PublicGitVCSConfig: &app.PublicGitVCSConfig{Repo: "acme/infra", Branch: "main"},
			},
		}
		build := &app.ComponentBuild{
			ComponentConfigConnection: app.ComponentConfigConnection{
				PublicGitVCSConfig: &app.PublicGitVCSConfig{Repo: "acme/infra", Branch: "main"},
			},
		}
		require.True(t, buildTracksBranchSource(run, build))
	})

	t.Run("same public repo and branch from component config", func(t *testing.T) {
		run := &app.AppBranchRun{
			AppBranchConfig: app.AppBranchConfig{
				PublicGitVCSConfig: &app.PublicGitVCSConfig{Repo: "acme/infra", Branch: "main"},
			},
		}
		build := &app.ComponentBuild{
			ComponentConfigConnection: app.ComponentConfigConnection{
				KubernetesManifestComponentConfig: &app.KubernetesManifestComponentConfig{
					PublicGitVCSConfig: &app.PublicGitVCSConfig{Repo: "acme/infra", Branch: "main"},
				},
			},
		}
		require.True(t, buildTracksBranchSource(run, build))
	})

	t.Run("mixed vcs types never match", func(t *testing.T) {
		run := &app.AppBranchRun{
			AppBranchConfig: app.AppBranchConfig{
				PublicGitVCSConfig: &app.PublicGitVCSConfig{Repo: "acme/infra", Branch: "main"},
			},
		}
		require.False(t, buildTracksBranchSource(run, connectedBuild("acme/infra", "main")))
	})

	t.Run("component with no vcs config", func(t *testing.T) {
		require.False(t, buildTracksBranchSource(
			connectedRun("acme/infra", "main"),
			&app.ComponentBuild{},
		))
	})
}

func TestBuildRefCanBePinned(t *testing.T) {
	t.Run("empty ref", func(t *testing.T) {
		require.True(t, buildRefCanBePinned(connectedBuild("acme/infra", "main")))
	})

	t.Run("public configured branch", func(t *testing.T) {
		ref := "main"
		require.True(t, buildRefCanBePinned(publicBuild("acme/infra", "main", &ref)))
	})

	t.Run("public explicit ref", func(t *testing.T) {
		ref := "feature"
		require.False(t, buildRefCanBePinned(publicBuild("acme/infra", "main", &ref)))
	})

	t.Run("connected explicit ref", func(t *testing.T) {
		ref := "feature"
		build := connectedBuild("acme/infra", "main")
		build.GitRef = &ref
		require.False(t, buildRefCanBePinned(build))
	})
}

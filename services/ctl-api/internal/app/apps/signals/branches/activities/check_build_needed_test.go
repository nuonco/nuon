package activities

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestBuildChangeReason(t *testing.T) {
	if got := buildChangeReason(true, "old", "new"); got != ChangeReasonSourceAndConfig {
		t.Fatalf("both: got %s", got)
	}
	if got := buildChangeReason(true, "same", "same"); got != ChangeReasonSourceChanged {
		t.Fatalf("source: got %s", got)
	}
	if got := buildChangeReason(false, "old", "new"); got != ChangeReasonConfigChanged {
		t.Fatalf("config: got %s", got)
	}
}

func TestReusableWithoutSourceResolution(t *testing.T) {
	imageCfg := func(tag, updatePolicy string) *app.ComponentConfigConnection {
		return &app.ComponentConfigConnection{
			ExternalImageComponentConfig: &app.ExternalImageComponentConfig{
				ImageURL:     "registry.example.com/acme/product",
				Tag:          tag,
				UpdatePolicy: updatePolicy,
			},
		}
	}
	helmRepoCfg := func(version string) *app.ComponentConfigConnection {
		return &app.ComponentConfigConnection{
			HelmComponentConfig: &app.HelmComponentConfig{
				HelmConfig: &app.HelmConfig{
					HelmRepoConfig: &app.HelmRepoConfig{
						RepoURL: "https://charts.example.com",
						Chart:   "example",
						Version: version,
					},
				},
			},
		}
	}
	imageBuild := &app.ComponentBuild{SourceDigest: "sha256:4f2b97443bf7980d327b132f6418063969db1cd037a92592352acbd0be7671d5"}
	helmBuild := &app.ComponentBuild{SourceChecksum: "671731285956d6d7eb9810acbeda7671798f6750c0b51a457fdd4c687c4b97ae"}

	tests := map[string]struct {
		cfg       *app.ComponentConfigConnection
		completed *app.ComponentBuild
		want      bool
	}{
		"image by tag with digest reuses": {
			cfg:       imageCfg("main-bazel-11cb9e3105558736968608182e27325fed7b396f", ""),
			completed: imageBuild,
			want:      true,
		},
		"image with update policy builds": {
			cfg:       imageCfg("1.2", "^1.2"),
			completed: imageBuild,
			want:      false,
		},
		"image without prior digest builds": {
			cfg:       imageCfg("main-bazel-11cb9e3105558736968608182e27325fed7b396f", ""),
			completed: &app.ComponentBuild{},
			want:      false,
		},
		"helm repo chart at exact version with checksum reuses": {
			cfg:       helmRepoCfg("0.173.1"),
			completed: helmBuild,
			want:      true,
		},
		"helm repo chart with version constraint builds": {
			cfg:       helmRepoCfg("~0.173.0"),
			completed: helmBuild,
			want:      false,
		},
		"helm repo chart with empty version builds": {
			cfg:       helmRepoCfg(""),
			completed: helmBuild,
			want:      false,
		},
		"helm repo chart without prior checksum builds": {
			cfg:       helmRepoCfg("0.173.1"),
			completed: &app.ComponentBuild{},
			want:      false,
		},
		"git sourced helm chart builds": {
			cfg: &app.ComponentConfigConnection{
				HelmComponentConfig: &app.HelmComponentConfig{
					ConnectedGithubVCSConfig: &app.ConnectedGithubVCSConfig{Repo: "acme/product"},
					HelmConfig: &app.HelmConfig{
						HelmRepoConfig: &app.HelmRepoConfig{Version: "0.173.1"},
					},
				},
			},
			completed: helmBuild,
			want:      false,
		},
		"git sourced component builds": {
			cfg: &app.ComponentConfigConnection{
				TerraformModuleComponentConfig: &app.TerraformModuleComponentConfig{
					ConnectedGithubVCSConfig: &app.ConnectedGithubVCSConfig{Repo: "acme/product"},
				},
			},
			completed: imageBuild,
			want:      false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := reusableWithoutSourceResolution(tc.cfg, tc.completed); got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

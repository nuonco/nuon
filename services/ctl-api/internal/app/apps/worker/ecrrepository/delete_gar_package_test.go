package ecrrepository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGARRepositoryURL(t *testing.T) {
	tests := map[string]struct {
		url     string
		want    *garRepository
		wantErr bool
	}{
		"valid": {
			url:  "us-central1-docker.pkg.dev/acme-project/nuon",
			want: &garRepository{location: "us-central1", project: "acme-project", name: "nuon"},
		},
		"trailing slash": {
			url:  "us-central1-docker.pkg.dev/acme-project/nuon/",
			want: &garRepository{location: "us-central1", project: "acme-project", name: "nuon"},
		},
		"multi-region": {
			url:  "us-docker.pkg.dev/acme-project/nuon",
			want: &garRepository{location: "us", project: "acme-project", name: "nuon"},
		},
		"empty":            {url: "", wantErr: true},
		"missing repo":     {url: "us-central1-docker.pkg.dev/acme-project", wantErr: true},
		"extra segment":    {url: "us-central1-docker.pkg.dev/acme-project/nuon/extra", wantErr: true},
		"non-gar host":     {url: "gcr.io/acme-project/nuon", wantErr: true},
		"host suffix only": {url: "-docker.pkg.dev/acme-project/nuon", wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := parseGARRepositoryURL(tt.url)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGARDeleteRequestPreservesEncodedSlash(t *testing.T) {
	repo := &garRepository{location: "us-central1", project: "acme-project", name: "nuon"}

	tests := map[string]struct {
		orgID    string
		appID    string
		wantPath string
	}{
		"org and app": {
			orgID:    "orgabc",
			appID:    "appdef",
			wantPath: "/v1/projects/acme-project/locations/us-central1/repositories/nuon/packages/orgabc%2Fappdef",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			req, err := newGARDeleteRequest(context.Background(), repo.packageName(tt.orgID, tt.appID))
			require.NoError(t, err)
			require.Equal(t, "artifactregistry.googleapis.com", req.URL.Host)
			require.Equal(t, tt.wantPath, req.URL.EscapedPath())
		})
	}
}

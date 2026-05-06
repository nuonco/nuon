package runnerimage

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseECRURL(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		wantRegistryID string
		wantRegion     string
		wantRepo       string
		wantErr        bool
	}{
		{
			name:           "simple repo",
			url:            "766121324316.dkr.ecr.us-west-2.amazonaws.com/runner",
			wantRegistryID: "766121324316",
			wantRegion:     "us-west-2",
			wantRepo:       "runner",
		},
		{
			name:           "nested repo path",
			url:            "766121324316.dkr.ecr.us-west-2.amazonaws.com/orgs/my-org/runner",
			wantRegistryID: "766121324316",
			wantRegion:     "us-west-2",
			wantRepo:       "orgs/my-org/runner",
		},
		{
			name:    "non-ecr url",
			url:     "ghcr.io/nuonco/runner",
			wantErr: true,
		},
		{
			name:    "empty",
			url:     "",
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id, region, repo, err := parseECRURL(tc.url)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantRegistryID, id)
			assert.Equal(t, tc.wantRegion, region)
			assert.Equal(t, tc.wantRepo, repo)
		})
	}
}

func TestIsGARURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"us-west1-docker.pkg.dev/proj/repo/runner", true},
		{"gcr.io/proj/runner", true},
		{"766121324316.dkr.ecr.us-west-2.amazonaws.com/runner", false},
		{"ghcr.io/foo/bar", false},
		{"", false},
	}
	for _, tc := range tests {
		t.Run(tc.url, func(t *testing.T) {
			assert.Equal(t, tc.want, isGARURL(tc.url))
		})
	}
}

type fakeECR struct {
	out *ecr.BatchGetImageOutput
	err error
}

func (f *fakeECR) BatchGetImage(_ context.Context, _ *ecr.BatchGetImageInput, _ ...func(*ecr.Options)) (*ecr.BatchGetImageOutput, error) {
	return f.out, f.err
}

func TestResolveWithClient(t *testing.T) {
	t.Run("returns digest on happy path", func(t *testing.T) {
		client := &fakeECR{out: &ecr.BatchGetImageOutput{
			Images: []ecrtypes.Image{
				{ImageId: &ecrtypes.ImageIdentifier{ImageDigest: aws.String("sha256:abc")}},
			},
		}}
		got, err := resolveWithClient(context.Background(), client, "123", "runner", "prod")
		require.NoError(t, err)
		assert.Equal(t, "sha256:abc", got)
	})
	t.Run("error when no images returned", func(t *testing.T) {
		client := &fakeECR{out: &ecr.BatchGetImageOutput{}}
		_, err := resolveWithClient(context.Background(), client, "123", "runner", "prod")
		assert.Error(t, err)
	})
	t.Run("error when ECR errors", func(t *testing.T) {
		client := &fakeECR{err: errors.New("boom")}
		_, err := resolveWithClient(context.Background(), client, "123", "runner", "prod")
		assert.Error(t, err)
	})
}

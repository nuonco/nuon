package containerimage

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"oras.land/oras-go/v2"

	"github.com/nuonco/nuon/pkg/oci/signature"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	pkgctx "github.com/nuonco/nuon/pkg/runner/ctx"
	"github.com/nuonco/nuon/pkg/sourceidentity"
	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const signedDigest = "sha256:96d02f455d5a73b817c0602910748609cf8471b1cc9522f78c75cedb1f67d072"

type fixedResolver struct{}

func (fixedResolver) Resolve(context.Context, *configs.OCIRegistryRepository, string) (*ocispec.Descriptor, error) {
	return &ocispec.Descriptor{Digest: signedDigest, MediaType: ocispec.MediaTypeImageIndex}, nil
}

func (fixedResolver) Tags(context.Context, *configs.OCIRegistryRepository) ([]string, error) {
	return nil, nil
}

type recordingCopier struct{ copied bool }

func (c *recordingCopier) Copy(context.Context, *configs.OCIRegistryRepository, string, *configs.OCIRegistryRepository, string) (*ocispec.Descriptor, error) {
	c.copied = true
	return &ocispec.Descriptor{}, nil
}

func (c *recordingCopier) CopyFromStore(context.Context, oras.ReadOnlyTarget, string, *configs.OCIRegistryRepository, string) (*ocispec.Descriptor, error) {
	return nil, nil
}

func (c *recordingCopier) CopyFromLocalRegistry(context.Context, string, *configs.OCIRegistryRepository, string) (*ocispec.Descriptor, error) {
	return nil, nil
}

type resultClient struct {
	nuonrunner.Client
	result *models.ServiceCreateRunnerJobExecutionResultRequest
}

func (c *resultClient) CreateJobExecutionResult(_ context.Context, _, _ string, req *models.ServiceCreateRunnerJobExecutionResultRequest) (*models.AppRunnerJobExecutionResult, error) {
	c.result = req
	return &models.AppRunnerJobExecutionResult{}, nil
}

func runSignedImageBuild(t *testing.T, subject string) (*models.ServiceCreateRunnerJobExecutionResultRequest, bool, error) {
	client := &resultClient{}

	copier := &recordingCopier{}
	h := &handler{
		apiClient:  client,
		ociCopy:    copier,
		ociResolve: fixedResolver{},
		state: &handlerState{
			jobID:          "job-1",
			jobExecutionID: "exec-1",
			resultTag:      "bld-1",
			plan:           &plantypes.BuildPlan{},
			regCfg:         &configs.OCIRegistryRepository{},
			cfg: &plantypes.ContainerImagePullPlan{
				Image: "cgr.dev/chainguard/static",
				Tag:   "latest",
				RepoCfg: &configs.OCIRegistryRepository{
					RegistryType: configs.OCIRegistryTypePublicOCI,
					Repository:   "cgr.dev/chainguard/static",
					OCIAuth:      &configs.OCIRegistryAuth{},
				},
				Verification: &signature.Verification{
					RequireSignature: true,
					Authorities: []signature.Authority{{
						Type:    signature.AuthorityTypeKeyless,
						Issuer:  "https://token.actions.githubusercontent.com",
						Subject: subject,
					}},
				},
			},
		},
	}

	ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
	err := h.Exec(ctx, &models.AppRunnerJob{ID: "job-1"}, &models.AppRunnerJobExecution{ID: "exec-1"})
	return client.result, copier.copied, err
}

func TestExecReportsSignatureOutcome(t *testing.T) {
	if os.Getenv("NUON_INTEGRATION") == "" {
		t.Skip("set NUON_INTEGRATION to run registry verification")
	}

	t.Run("trusted authority", func(t *testing.T) {
		result, copied, err := runSignedImageBuild(t, "https://github.com/chainguard-images/images/.github/workflows/release.yaml@refs/heads/main")
		require.NoError(t, err)
		require.True(t, copied)
		require.True(t, result.Success)
		require.Equal(t, signedDigest, result.SourceDigest)
	})

	t.Run("untrusted authority", func(t *testing.T) {
		result, copied, err := runSignedImageBuild(t, "https://github.com/acme/images/.github/workflows/release.yaml@refs/heads/main")
		require.Error(t, err)
		require.False(t, copied)
		require.False(t, result.Success)
		require.Equal(t, signature.VerifyStep, result.ErrorMetadata["step"])
	})
}

type tagResolver map[string]string

func (r tagResolver) Resolve(_ context.Context, _ *configs.OCIRegistryRepository, ref string) (*ocispec.Descriptor, error) {
	if strings.HasPrefix(ref, "sha256:") {
		return &ocispec.Descriptor{Digest: digest.Digest(ref), MediaType: ocispec.MediaTypeImageIndex}, nil
	}
	return &ocispec.Descriptor{Digest: digest.Digest(r[ref]), MediaType: ocispec.MediaTypeImageIndex}, nil
}

func (r tagResolver) Tags(context.Context, *configs.OCIRegistryRepository) ([]string, error) {
	tags := make([]string, 0, len(r))
	for tag := range r {
		tags = append(tags, tag)
	}
	return tags, nil
}

func TestExecResumedUpdatePolicyBuildReportsSelectedTag(t *testing.T) {
	const claimed = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	const newer = "sha256:2222222222222222222222222222222222222222222222222222222222222222"
	for name, tc := range map[string]struct {
		tags    tagResolver
		wantTag string
	}{
		"selected tag still points at the claimed digest": {tags: tagResolver{"1.25.4": newer, "1.25.5": claimed}, wantTag: "1.25.5"},
		"a newer tag was published since the claim":       {tags: tagResolver{"1.25.5": claimed, "1.25.6": newer}, wantTag: ""},
	} {
		t.Run(name, func(t *testing.T) {
			identity := sourceidentity.New("image", "config")
			identity.Digest, identity.Captured = claimed, true
			identity.Seal()
			client := &resultClient{}
			h := &handler{
				apiClient:  client,
				ociCopy:    &recordingCopier{},
				ociResolve: tc.tags,
				state: &handlerState{
					jobID:          "job-1",
					jobExecutionID: "exec-1",
					resultTag:      "bld-1",
					plan:           &plantypes.BuildPlan{SourceIdentity: identity},
					regCfg:         &configs.OCIRegistryRepository{},
					cfg: &plantypes.ContainerImagePullPlan{
						Image:        "ghcr.io/acme/api",
						UpdatePolicy: "~1.25",
						RepoCfg: &configs.OCIRegistryRepository{
							RegistryType: configs.OCIRegistryTypePublicOCI,
							Repository:   "ghcr.io/acme/api",
							OCIAuth:      &configs.OCIRegistryAuth{},
						},
					},
				},
			}
			ctx := pkgctx.SetLogger(context.Background(), zap.NewNop())
			require.NoError(t, h.Exec(ctx, &models.AppRunnerJob{ID: "job-1"}, &models.AppRunnerJobExecution{ID: "exec-1"}))
			require.Equal(t, claimed, client.result.SourceDigest)
			require.Equal(t, tc.wantTag, client.result.ResolvedTag)
		})
	}
}

package runnerimage

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
)

// ecrURLPattern matches `<account>.dkr.ecr.<region>.amazonaws.com/<repo>` and
// captures the account ID, region, and repository name. The repository part may
// include nested paths (e.g. `org-id/app-id`).
var ecrURLPattern = regexp.MustCompile(`^([0-9]+)\.dkr\.ecr\.([a-z0-9-]+)\.amazonaws\.com/(.+)$`)

type ecrAPI interface {
	BatchGetImage(ctx context.Context, params *ecr.BatchGetImageInput, optFns ...func(*ecr.Options)) (*ecr.BatchGetImageOutput, error)
}

// ResolveAWSImageDigest looks up the immutable manifest digest for an
// `imageURL:tag` pair in our management-account ECR. Returns digests in the
// `sha256:...` form. The caller is responsible for treating empty results as
// "not pinned" and proceeding with the mutable tag.
func ResolveAWSImageDigest(ctx context.Context, cfg *internal.Config, imageURL, tag string) (string, error) {
	registryID, region, repo, err := parseECRURL(imageURL)
	if err != nil {
		return "", err
	}

	awsCfg, err := credentials.Fetch(ctx, &credentials.Config{
		Region: region,
		AssumeRole: &credentials.AssumeRoleConfig{
			RoleARN:     cfg.ManagementIAMRoleARN,
			SessionName: "ctl-api-runner-image-pin",
		},
	})
	if err != nil {
		return "", fmt.Errorf("unable to fetch credentials: %w", err)
	}

	return resolveWithClient(ctx, ecr.NewFromConfig(awsCfg), registryID, repo, tag)
}

func resolveWithClient(ctx context.Context, client ecrAPI, registryID, repo, tag string) (string, error) {
	out, err := client.BatchGetImage(ctx, &ecr.BatchGetImageInput{
		RegistryId:     aws.String(registryID),
		RepositoryName: aws.String(repo),
		ImageIds: []ecrtypes.ImageIdentifier{
			{ImageTag: generics.ToPtr(tag)},
		},
	})
	if err != nil {
		return "", fmt.Errorf("unable to fetch image %s/%s:%s: %w", registryID, repo, tag, err)
	}
	if len(out.Images) == 0 {
		return "", fmt.Errorf("no images returned for %s/%s:%s", registryID, repo, tag)
	}
	digest := aws.ToString(out.Images[0].ImageId.ImageDigest)
	if digest == "" {
		return "", fmt.Errorf("empty digest for %s/%s:%s", registryID, repo, tag)
	}
	return digest, nil
}

func parseECRURL(imageURL string) (registryID, region, repo string, err error) {
	m := ecrURLPattern.FindStringSubmatch(strings.TrimSpace(imageURL))
	if m == nil {
		return "", "", "", fmt.Errorf("not an ECR image url: %s", imageURL)
	}
	return m[1], m[2], m[3], nil
}

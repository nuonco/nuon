package prune

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/smithy-go"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

const ecrBatchSize = 100

var acceptedManifestMediaTypes = []string{
	"application/vnd.oci.image.index.v1+json",
	"application/vnd.oci.image.manifest.v1+json",
	"application/vnd.docker.distribution.manifest.list.v2+json",
	"application/vnd.docker.distribution.manifest.v2+json",
	"application/vnd.docker.distribution.manifest.v1+json",
}

var ErrPermissionDenied = errors.New("permission denied")

type ECRRepository struct {
	client     *ecr.Client
	registryID string
	name       string
	URL        string
}

func NewECRRepository(ctx context.Context, cfg *configs.OCIRegistryRepository) (*ECRRepository, error) {
	if cfg.RegistryType != configs.OCIRegistryTypeECR {
		return nil, fmt.Errorf("unsupported registry type %q", cfg.RegistryType)
	}
	if cfg.ECRAuth == nil {
		return nil, errors.New("missing ecr auth")
	}

	url, registryID, name, err := parseECRRepository(cfg.Repository)
	if err != nil {
		return nil, err
	}

	awsCfg, err := awscredentials.Fetch(ctx, cfg.ECRAuth)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch aws credentials: %w", err)
	}
	if cfg.Region != "" {
		awsCfg.Region = cfg.Region
	}

	return &ECRRepository{
		client:     ecr.NewFromConfig(awsCfg),
		registryID: registryID,
		name:       name,
		URL:        url,
	}, nil
}

func parseECRRepository(repository string) (string, string, string, error) {
	url := strings.TrimPrefix(strings.TrimPrefix(repository, "https://"), "http://")
	host, name, ok := strings.Cut(url, "/")
	if !ok || name == "" {
		return "", "", "", fmt.Errorf("invalid ecr repository %q", repository)
	}
	registryID, _, ok := strings.Cut(host, ".")
	if !ok || registryID == "" {
		return "", "", "", fmt.Errorf("invalid ecr registry host %q", host)
	}
	return url, registryID, name, nil
}

func (r *ECRRepository) ListManifests(ctx context.Context) ([]Manifest, error) {
	var manifests []Manifest
	pager := ecr.NewDescribeImagesPaginator(r.client, &ecr.DescribeImagesInput{
		RegistryId:     aws.String(r.registryID),
		RepositoryName: aws.String(r.name),
	})
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, wrapAWSErr("describe images", err)
		}
		for _, img := range page.ImageDetails {
			m := Manifest{
				Digest: aws.ToString(img.ImageDigest),
				Tags:   img.ImageTags,
			}
			if img.ImagePushedAt != nil {
				m.PushedAt = *img.ImagePushedAt
			}
			manifests = append(manifests, m)
		}
	}

	if err := r.fillReferences(ctx, manifests); err != nil {
		return nil, err
	}
	return manifests, nil
}

type manifestRefs struct {
	Manifests []struct {
		Digest string `json:"digest"`
	} `json:"manifests"`
	Subject *struct {
		Digest string `json:"digest"`
	} `json:"subject"`
}

func (r *ECRRepository) fillReferences(ctx context.Context, manifests []Manifest) error {
	idx := make(map[string]int, len(manifests))
	for i, m := range manifests {
		idx[m.Digest] = i
	}

	for start := 0; start < len(manifests); start += ecrBatchSize {
		end := min(start+ecrBatchSize, len(manifests))
		ids := make([]ecrtypes.ImageIdentifier, 0, end-start)
		for _, m := range manifests[start:end] {
			ids = append(ids, ecrtypes.ImageIdentifier{ImageDigest: aws.String(m.Digest)})
		}

		out, err := r.client.BatchGetImage(ctx, &ecr.BatchGetImageInput{
			RegistryId:         aws.String(r.registryID),
			RepositoryName:     aws.String(r.name),
			ImageIds:           ids,
			AcceptedMediaTypes: acceptedManifestMediaTypes,
		})
		if err != nil {
			return wrapAWSErr("batch get image", err)
		}
		if len(out.Failures) > 0 {
			f := out.Failures[0]
			return fmt.Errorf("unable to fetch manifest %s: %s: %s",
				aws.ToString(f.ImageId.ImageDigest), f.FailureCode, aws.ToString(f.FailureReason))
		}

		fetched := map[string]bool{}
		for _, img := range out.Images {
			digest := aws.ToString(img.ImageId.ImageDigest)
			if fetched[digest] {
				continue
			}
			fetched[digest] = true

			i, ok := idx[digest]
			if !ok {
				continue
			}
			var refs manifestRefs
			if err := json.Unmarshal([]byte(aws.ToString(img.ImageManifest)), &refs); err != nil {
				return fmt.Errorf("unable to parse manifest %s: %w", digest, err)
			}
			for _, child := range refs.Manifests {
				manifests[i].Children = append(manifests[i].Children, child.Digest)
			}
			if refs.Subject != nil {
				manifests[i].Subject = refs.Subject.Digest
			}
		}
		for _, m := range manifests[start:end] {
			if !fetched[m.Digest] {
				return fmt.Errorf("manifest %s was not returned", m.Digest)
			}
		}
	}
	return nil
}

type DeleteFailure struct {
	Digest string `json:"digest"`
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

func (r *ECRRepository) Delete(ctx context.Context, deletions []Deletion) ([]string, []DeleteFailure, error) {
	var deleted []string
	var failures []DeleteFailure
	for start := 0; start < len(deletions); start += ecrBatchSize {
		end := min(start+ecrBatchSize, len(deletions))
		ids := make([]ecrtypes.ImageIdentifier, 0, end-start)
		for _, d := range deletions[start:end] {
			ids = append(ids, ecrtypes.ImageIdentifier{ImageDigest: aws.String(d.Digest)})
		}

		out, err := r.client.BatchDeleteImage(ctx, &ecr.BatchDeleteImageInput{
			RegistryId:     aws.String(r.registryID),
			RepositoryName: aws.String(r.name),
			ImageIds:       ids,
		})
		if err != nil {
			return deleted, failures, wrapAWSErr("batch delete image", err)
		}
		seen := map[string]bool{}
		for _, id := range out.ImageIds {
			digest := aws.ToString(id.ImageDigest)
			if !seen[digest] {
				seen[digest] = true
				deleted = append(deleted, digest)
			}
		}
		for _, f := range out.Failures {
			failures = append(failures, DeleteFailure{
				Digest: aws.ToString(f.ImageId.ImageDigest),
				Code:   string(f.FailureCode),
				Reason: aws.ToString(f.FailureReason),
			})
		}
	}
	return deleted, failures, nil
}

func wrapAWSErr(op string, err error) error {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) && apiErr.ErrorCode() == "AccessDeniedException" {
		return fmt.Errorf("%s: %w: %s", op, ErrPermissionDenied, apiErr.ErrorMessage())
	}
	return fmt.Errorf("%s: %w", op, err)
}

func (r *ECRRepository) RepositoryURL() string {
	return r.URL
}

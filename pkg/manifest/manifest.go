package manifest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/aws/s3downloader"
	"github.com/powertoolsdev/mono/pkg/aws/s3uploader"
)

const (
	ManifestBucketName = "nuon-build-manifests"
	ManifestRegion     = "us-west-2"
)

// BuildManifest represents the complete manifest for a build
type BuildManifest struct {
	Service   ServiceInfo    `json:"service"`
	Artifacts []ArtifactInfo `json:"artifacts,omitempty"`
	Metadata  BuildMetadata  `json:"metadata"`
	Timestamp time.Time      `json:"timestamp"`
}

// ServiceInfo contains details about the service image
type ServiceInfo struct {
	ImageURL    string            `json:"image_url,omitempty"`
	ImageDigest string            `json:"image_digest,omitempty"`
	ImageSize   int64             `json:"image_size,omitempty"`
	ECR         ECRInfo           `json:"ecr,omitempty"`
	BuildArgs   map[string]string `json:"build_args,omitempty"`
}

// ECRInfo contains ECR repository information
type ECRInfo struct {
	RepositoryURL string `json:"repository_url,omitempty"`
	RepositoryARN string `json:"repository_arn,omitempty"`
	Registry      string `json:"registry,omitempty"`
	Tag           string `json:"tag,omitempty"`
}

// ArtifactInfo contains details about artifacts (S3 files)
type ArtifactInfo struct {
	S3Path   string  `json:"s3_path,omitempty"`
	Checksum string  `json:"checksum,omitempty"`
	Size     int64   `json:"size,omitempty"`
	ECR      ECRInfo `json:"ecr,omitempty"`
}

// BuildMetadata contains metadata about the build
type BuildMetadata struct {
	IsPromotion bool    `json:"is_promotion"`
	PRDetails   *PRInfo `json:"pr_details,omitempty"`
	IsMain      bool    `json:"is_main"`
	Author      string  `json:"author,omitempty"` // Workflow trigger (may differ from PR author)
	GitRef      string  `json:"git_ref,omitempty"`
	GitCommit   string  `json:"git_commit,omitempty"`
}

// PRInfo contains pull request information
type PRInfo struct {
	Number     string `json:"number"`
	Title      string `json:"title,omitempty"`
	Author     string `json:"author,omitempty"`
	HeadBranch string `json:"head_branch,omitempty"`
	BaseBranch string `json:"base_branch,omitempty"`
	URL        string `json:"url,omitempty"`
}

// Client handles manifest generation and upload
type Client struct {
	cfg      *Config
	uploader s3uploader.Uploader
}

// Config contains configuration for creating a manifest client
type Config struct {
	IsCI              bool
	RoleARN           string
	DisableGithubOIDC bool
}

// New creates a new manifest client
func New(config *Config) (*Client, error) {
	// Guard against nil config
	if config == nil {
		config = &Config{}
	}

	cfg := &credentials.Config{
		Region: ManifestRegion,
	}
	if config.IsCI {
		cfg.AssumeRole = &credentials.AssumeRoleConfig{
			RoleARN:       config.RoleARN,
			SessionName:   "nuonctl",
			UseGithubOIDC: !config.DisableGithubOIDC,
		}
	} else {
		cfg.Profile = "infra-shared-prod.NuonAdmin"
	}

	uploader, err := s3uploader.NewS3Uploader(validator.New(),
		s3uploader.WithBucketName(ManifestBucketName),
		s3uploader.WithCredentials(cfg))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create s3 uploader")
	}

	return &Client{
		cfg:      config,
		uploader: uploader,
	}, nil
}

// GenerateManifest creates a build manifest
func GenerateManifest(
	imageURL string,
	ecrInfo ECRInfo,
	buildArgs map[string]string,
	metadata BuildMetadata,
) *BuildManifest {
	return &BuildManifest{
		Service: ServiceInfo{
			ImageURL:  imageURL,
			ECR:       ecrInfo,
			BuildArgs: buildArgs,
		},
		Artifacts: []ArtifactInfo{},
		Metadata:  metadata,
		Timestamp: time.Now().UTC(),
	}
}

// AddArtifact adds an artifact to the manifest
func (m *BuildManifest) AddArtifact(artifact ArtifactInfo) {
	m.Artifacts = append(m.Artifacts, artifact)
}

// SetImageDetails updates the service image details (digest and size)
func (m *BuildManifest) SetImageDetails(digest string, size int64) {
	m.Service.ImageDigest = digest
	m.Service.ImageSize = size
}

// ToJSON converts the manifest to JSON
func (m *BuildManifest) ToJSON() ([]byte, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal manifest to JSON")
	}
	return data, nil
}

// Upload uploads the manifest to S3
func (c *Client) Upload(ctx context.Context, manifest *BuildManifest, serviceName, version string) error {
	// Generate manifest JSON
	manifestJSON, err := manifest.ToJSON()
	if err != nil {
		return errors.Wrap(err, "unable to generate manifest JSON")
	}

	// Create S3 key: <service>/<version>/manifest.json
	s3Key := fmt.Sprintf("%s/%s/manifest.json", serviceName, version)

	// Upload to S3
	if err := c.uploader.UploadBlob(ctx, manifestJSON, s3Key); err != nil {
		return errors.Wrapf(err, "unable to upload manifest to s3://%s/%s", ManifestBucketName, s3Key)
	}

	return nil
}

// Download downloads and parses a manifest from S3
func (c *Client) Download(ctx context.Context, serviceName, version string) (*BuildManifest, error) {
	// Create S3 key: <service>/<version>/manifest.json
	s3Key := fmt.Sprintf("%s/%s/manifest.json", serviceName, version)

	// Create credentials config - mirror upload behavior
	credCfg := &credentials.Config{
		Region: ManifestRegion,
	}
	if c.cfg.IsCI {
		credCfg.AssumeRole = &credentials.AssumeRoleConfig{
			RoleARN:       c.cfg.RoleARN,
			SessionName:   "nuonctl",
			UseGithubOIDC: !c.cfg.DisableGithubOIDC,
		}
	} else {
		credCfg.Profile = "infra-shared-prod.NuonAdmin"
	}

	// Create S3 downloader
	downloader, err := s3downloader.New(ManifestBucketName, s3downloader.WithCredentials(credCfg))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create s3 downloader")
	}

	// Download manifest JSON
	data, err := downloader.GetBlob(ctx, s3Key)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to download manifest from s3://%s/%s", ManifestBucketName, s3Key)
	}

	// Parse JSON
	var manifest BuildManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, errors.Wrap(err, "unable to parse manifest JSON")
	}

	return &manifest, nil
}

// CompareImageDigests compares local and remote image digests
// Returns: hasUpdate, localDigest, remoteDigest, error
// If error is non-nil, comparison failed and caller should handle gracefully
func (c *Client) CompareImageDigests(ctx context.Context, serviceName, tag, imageURL string, getLocalInfo func(context.Context, string) (*DockerInspectOutput, error)) (bool, string, string, error) {
	// Try to download the manifest for this tag
	remoteManifest, err := c.Download(ctx, serviceName, tag)
	if err != nil {
		return false, "", "", err
	}
	if remoteManifest == nil || remoteManifest.Service.ImageDigest == "" {
		return false, "", "", errors.New("remote manifest missing or has no image digest")
	}

	// Check local image digest
	localInfo, err := getLocalInfo(ctx, imageURL)
	if err != nil {
		return false, "", "", err
	}
	if localInfo == nil || len(localInfo.RepoDigests) == 0 {
		return false, "", "", errors.New("local image not found or has no digest")
	}

	// Extract local digest
	localDigest := ""
	for _, repoDigest := range localInfo.RepoDigests {
		parts := strings.Split(repoDigest, "@")
		if len(parts) == 2 {
			localDigest = parts[1]
			break
		}
	}

	if localDigest == "" {
		return false, "", "", errors.New("failed to extract digest from local image repo digests")
	}

	// Compare digests
	hasUpdate := localDigest != remoteManifest.Service.ImageDigest
	return hasUpdate, localDigest, remoteManifest.Service.ImageDigest, nil
}

// DockerInspectOutput represents the docker inspect output format
type DockerInspectOutput struct {
	ID          string   `json:"Id"`
	RepoDigests []string `json:"RepoDigests"`
}

// GetMetadataFromEnv extracts build metadata from environment variables
//
// Standard GitHub Actions variables:
//   - GITHUB_ACTOR: GitHub username of the person who triggered the workflow
//   - GITHUB_REF: Full git ref (e.g., refs/heads/main, refs/pull/123/merge)
//   - GITHUB_SHA: Commit SHA that triggered the workflow
//   - GITHUB_REPOSITORY: Repository name (owner/repo)
//   - GITHUB_HEAD_REF: Source branch of PR (only for pull_request events)
//   - GITHUB_BASE_REF: Target branch of PR (only for pull_request events)
//
// Custom variables:
//   - CI_ACTION_REF_NAME_SLUG: Branch slug (from FranzDiebold action)
//   - CI_PR_NUMBER: PR number (from FranzDiebold action)
//   - CI_PR_TITLE: PR title (from FranzDiebold action)
//   - BUILD_PROMOTION: "true" if this is a promotion build
//   - GITHUB_PR_AUTHOR: PR author (set manually from github.event.pull_request.user.login)
//   - GIT_COMMIT: Fallback commit SHA if GITHUB_SHA is not set
func GetMetadataFromEnv() BuildMetadata {
	metadata := BuildMetadata{
		IsMain:      os.Getenv("CI_ACTION_REF_NAME_SLUG") == "main",
		Author:      os.Getenv("GITHUB_ACTOR"),
		GitRef:      os.Getenv("GITHUB_REF"),
		GitCommit:   os.Getenv("GITHUB_SHA"),
		IsPromotion: os.Getenv("BUILD_PROMOTION") == "true",
	}

	// If GITHUB_SHA is empty, try to get commit from git
	if metadata.GitCommit == "" {
		metadata.GitCommit = os.Getenv("GIT_COMMIT")
	}

	// Extract PR details if available (non-blocking, never fails)
	if prNum := os.Getenv("CI_PR_NUMBER"); prNum != "" {
		// Build detailed PR info only if we have at least the PR number
		prInfo := &PRInfo{
			Number: prNum,
		}

		// Add optional fields if available
		if title := os.Getenv("CI_PR_TITLE"); title != "" {
			prInfo.Title = title
		}
		if author := os.Getenv("GITHUB_PR_AUTHOR"); author != "" {
			prInfo.Author = author
		}
		if headBranch := os.Getenv("GITHUB_HEAD_REF"); headBranch != "" {
			prInfo.HeadBranch = headBranch
		}
		if baseBranch := os.Getenv("GITHUB_BASE_REF"); baseBranch != "" {
			prInfo.BaseBranch = baseBranch
		}

		// Construct PR URL if we have repo info (best effort)
		if repo := os.Getenv("GITHUB_REPOSITORY"); repo != "" {
			prInfo.URL = fmt.Sprintf("https://github.com/%s/pull/%s", repo, prNum)
		}

		metadata.PRDetails = prInfo
	}

	return metadata
}

type ManifestWithVersion struct {
	Version  string
	Manifest *BuildManifest
}

type DownloadOptions struct {
	PromotionsOnly bool
}

func (c *Client) DownloadManifestsParallel(ctx context.Context, serviceName string, versions []string, opts *DownloadOptions) []*ManifestWithVersion {
	if opts == nil {
		opts = &DownloadOptions{}
	}

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		results = make([]*ManifestWithVersion, 0)
	)

	for _, version := range versions {
		wg.Add(1)
		go func(ver string) {
			defer wg.Done()

			buildManifest, err := c.Download(ctx, serviceName, ver)
			if err != nil {
				return
			}

			if opts.PromotionsOnly && !buildManifest.Metadata.IsPromotion {
				return
			}

			mu.Lock()
			results = append(results, &ManifestWithVersion{
				Version:  ver,
				Manifest: buildManifest,
			})
			mu.Unlock()
		}(version)
	}

	wg.Wait()
	return results
}

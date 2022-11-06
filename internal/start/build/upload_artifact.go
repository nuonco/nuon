package build

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/mitchellh/mapstructure"
	"github.com/powertoolsdev/go-uploader"
	"google.golang.org/grpc"
)

const artifactFilename string = "artifact.json"

type UploadArtifactRequest struct {
	DeploymentID  string `json:"deployment_id" validate:"required"`
	ComponentName string `json:"component_name" validate:"required"`

	OrgID                string `json:"org_id" validate:"required"`
	AppID                string `json:"app_id" validate:"required"`
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`

	BucketName   string `json:"bucket_name" validate:"required"`
	BucketPrefix string `json:"bucket_prefix" validate:"required"`
}

func (u UploadArtifactRequest) validate() error {
	validate := validator.New()
	return validate.Struct(u)
}

type UploadArtifactResponse struct{}

func (a *Activities) UploadArtifact(ctx context.Context, req UploadArtifactRequest) (UploadArtifactResponse, error) {
	var resp UploadArtifactResponse
	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	client, err := a.GetOrgWaypointClient(ctx, req.TokenSecretNamespace, req.OrgID, req.OrgServerAddr)
	if err != nil {
		return resp, fmt.Errorf("unable to get org waypoint client: %w", err)
	}

	build, err := a.getWaypointBuild(ctx, client, req)
	if err != nil {
		return resp, fmt.Errorf("unable get org waypoint build: %w", err)
	}

	uploadClient := uploader.NewS3Uploader(req.BucketName, req.BucketPrefix)
	if err := a.uploadArtifactMetadata(ctx, uploadClient, build, req); err != nil {
		return resp, fmt.Errorf("unable to upload artifact metadata: %s", err)
	}

	return resp, nil
}

type artifactUploader interface {
	getWaypointBuild(context.Context, waypointClientBuildLister, UploadArtifactRequest) (*gen.Build, error)
	uploadArtifactMetadata(context.Context, s3BlobUploader, *gen.Build, UploadArtifactRequest) error
}

type waypointClientBuildLister interface {
	ListBuilds(ctx context.Context, in *gen.ListBuildsRequest, opts ...grpc.CallOption) (*gen.ListBuildsResponse, error)
}

type artifactUploaderImpl struct{}

type artifactJSON struct {
	Request  UploadArtifactRequest  `json:"request"`
	Artifact map[string]interface{} `json:"artifact"`
	Build    *gen.Build             `json:"build"`
}

type s3BlobUploader interface {
	UploadBlob(context.Context, []byte, string) error
}

func (artifactUploaderImpl) uploadArtifactMetadata(ctx context.Context, client s3BlobUploader, build *gen.Build, req UploadArtifactRequest) error {
	var buildArtifact map[string]interface{}
	if err := mapstructure.Decode(build.Artifact, &buildArtifact); err != nil {
		return fmt.Errorf("unable to convert build artifact json")
	}

	art := artifactJSON{
		Build:    build,
		Request:  req,
		Artifact: buildArtifact,
	}
	byts, err := json.Marshal(art)
	if err != nil {
		return fmt.Errorf("unable to create json from artifact: %s", art)
	}

	if err := client.UploadBlob(ctx, byts, artifactFilename); err != nil {
		return fmt.Errorf("unable to upload to s3: %s", err)
	}

	return nil
}

var errBuildNotFound = fmt.Errorf("build not found")

func (artifactUploaderImpl) getWaypointBuild(ctx context.Context, client waypointClientBuildLister, req UploadArtifactRequest) (*gen.Build, error) {
	bReq := &gen.ListBuildsRequest{
		Application: &gen.Ref_Application{
			Application: req.ComponentName,
			Project:     req.OrgID,
		},
		Workspace: &gen.Ref_Workspace{
			Workspace: req.AppID,
		},
		Order: &gen.OperationOrder{
			Desc: true,
		},
	}
	resp, err := client.ListBuilds(ctx, bReq)
	if err != nil {
		return nil, err
	}

	for _, build := range resp.Builds {
		deploymentID, ok := build.Labels["deployment-id"]
		if !ok {
			continue
		}
		if deploymentID == req.DeploymentID {
			return build, nil
		}
	}

	return nil, errBuildNotFound
}

var _ artifactUploader = (*artifactUploaderImpl)(nil)

package provision

import (
	"context"

	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/go-uploader"
	"github.com/powertoolsdev/go-waypoint"
)

type GenerateWaypointConfigRequest struct {
	ProvisionRequest

	BucketName   string `json:"bucket_name"   validate:"required"`
	BucketPrefix string `json:"bucket_prefix" validate:"required"`
}

func (p GenerateWaypointConfigRequest) validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

type GenerateWaypointConfigResponse struct{}

// generator exposes the methods to properly generate a waypoint config
type waypointCfgGenerator interface {
	generateWaypointCfg(context.Context, GenerateWaypointConfigRequest) error
}

var _ waypointCfgGenerator = (*waypointCfgGeneratorImpl)(nil)

type waypointCfgGeneratorImpl struct{}

func (a *Activities) GenerateWaypointConfig(
	ctx context.Context,
	req GenerateWaypointConfigRequest,
) (GenerateWaypointConfigResponse, error) {
	resp := GenerateWaypointConfigResponse{}
	if err := req.validate(); err != nil {
		return resp, err
	}

	err := a.generateWaypointCfg(ctx, req)
	if err != nil {
		return resp, nil
	}

	return resp, nil
}

func (w *waypointCfgGeneratorImpl) generateWaypointCfg(ctx context.Context, req GenerateWaypointConfigRequest) error {
	component := req.Component
	config := component.GenerateHCL(waypoint.HCLDeploy)

	s3upload := uploader.NewS3Uploader(req.BucketName, req.BucketPrefix)

	err := s3upload.UploadBlob(ctx, config.Bytes(), "deploy.hcl")
	if err != nil {
		return err
	}

	return nil
}

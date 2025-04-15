package activities

import "context"

type UpdateAWSCloudFormationStackVersionRequest struct {
	ID string `validate:"required"`

	Contents           []byte `validate:"required"`
	AWSBucketName      string `validate:"required"`
	AWSBucketKey       string `validate:"required"`
	QuickLinkPublicURL string `validate:"required"`
}

func (a *Activities) UpdateAWSCloudFormationStackVersion(ctx context.Context, req *UpdateAWSCloudFormationStackVersionRequest) error {
	return nil
}

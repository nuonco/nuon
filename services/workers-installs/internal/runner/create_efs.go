package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type CreateEFSRequest struct {
	IAMRoleARN string `validate:"required"`
	InstallID  string `validate:"required"`
	Region     string `validate:"required"`
}

type CreateEFSResponse struct{}

func (a *Activities) CreateEFS(ctx context.Context, req *CreateEFSRequest) (*CreateEFSResponse, error) {
	efsClient, err := a.getEFSClient(ctx, req.IAMRoleARN, req.Region)
	if err != nil {
		return nil, fmt.Errorf("unable to get efs service: %w", err)
	}

	_, err = efsClient.CreateFileSystem(ctx, &efs.CreateFileSystemInput{
		CreationToken: generics.ToPtr(req.InstallID),
		Encrypted:     generics.ToPtr(true),
		Tags: []efstypes.Tag{
			{
				Key:   generics.ToPtr(defaultRunnerTagName),
				Value: generics.ToPtr(defaultRunnerTagValue),
			},
			{
				Key:   generics.ToPtr(defaultRunnerIDTagName),
				Value: generics.ToPtr(req.InstallID),
			},
		},
	})
	if err != nil {
		alreadyExistsErr := &efstypes.FileSystemAlreadyExists{}
		if errors.As(err, &alreadyExistsErr); err != nil {
			return &CreateEFSResponse{}, nil
		}

		return nil, fmt.Errorf("unable to create file system: %w", err)
	}

	return &CreateEFSResponse{}, nil
}

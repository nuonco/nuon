package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"

	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
)

type DeleteEFSRequest struct {
	IAMRoleARN string `validate:"required"`
	InstallID  string `validate:"required"`
	Region     string `validate:"required"`

	TwoStepConfig *assumerole.TwoStepConfig `validate:"required"`
}

type DeleteEFSResponse struct{}

func (a *Activities) DeleteEFS(ctx context.Context, req *DeleteEFSRequest) (*DeleteEFSResponse, error) {
	efsClient, err := a.getEFSClient(ctx, req.IAMRoleARN, req.Region, req.TwoStepConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to get efs service: %w", err)
	}

	fs, err := a.getEFS(ctx, efsClient, req.InstallID)
	nfe := &efstypes.FileSystemNotFound{}
	if errors.As(err, &nfe) {
		return &DeleteEFSResponse{}, nil
	}

	_, err = efsClient.DeleteFileSystem(ctx, &efs.DeleteFileSystemInput{
		FileSystemId: fs.FileSystemId,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to delete file system: %w", err)
	}

	return &DeleteEFSResponse{}, nil
}

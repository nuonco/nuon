package runner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type PollEFSRequest struct {
	IAMRoleARN string `validate:"required"`
	InstallID  string `validate:"required"`
}

type PollEFSResponse struct {
	FsID string
}

func (a *Activities) PollEFS(ctx context.Context, req PollEFSRequest) (*PollEFSResponse, error) {
	efsClient, err := a.getEFSClient(ctx, req.IAMRoleARN)
	if err != nil {
		return nil, fmt.Errorf("unable to create efs client: %w", err)
	}

	fss, err := efsClient.DescribeFileSystems(ctx, &efs.DescribeFileSystemsInput{
		CreationToken: generics.ToPtr(req.IAMRoleARN),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get efs: %w", err)
	}

	if len(fss.FileSystems) != 1 {
		return nil, fmt.Errorf("unable to find efs for install: %w", err)
	}

	fs := fss.FileSystems[0]
	if fs.LifeCycleState != efstypes.LifeCycleStateAvailable {
		return nil, fmt.Errorf("efs is not in valid life cycle state: %w", err)
	}

	return &PollEFSResponse{
		FsID: *fs.FileSystemId,
	}, nil
}

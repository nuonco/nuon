package runner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type PollEFSMountTargetsRequest struct {
	IAMRoleARN string `validate:"required"`
	FsID       string `validate:"required"`
	Region     string `validate:"required"`
}

type PollEFSMountTargetsResponse struct {
	FsID string
}

func (a *Activities) PollEFSMountTargets(ctx context.Context, req PollEFSMountTargetsRequest) (*PollEFSMountTargetsResponse, error) {
	efsClient, err := a.getEFSClient(ctx, req.IAMRoleARN, req.Region)
	if err != nil {
		return nil, fmt.Errorf("unable to create efs client: %w", err)
	}

	mountTargets, err := efsClient.DescribeMountTargets(ctx, &efs.DescribeMountTargetsInput{
		FileSystemId: generics.ToPtr(req.FsID),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get efs: %w", err)
	}

	for _, mountTarget := range mountTargets.MountTargets {
		if mountTarget.LifeCycleState != efstypes.LifeCycleStateAvailable {
			return nil, fmt.Errorf("mount target is not available: %s", mountTarget.LifeCycleState)
		}
	}

	return &PollEFSMountTargetsResponse{}, nil
}

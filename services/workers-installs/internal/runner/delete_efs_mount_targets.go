package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"

	assumerole "github.com/powertoolsdev/mono/pkg/aws/assume-role"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type DeleteEFSMountTargetsRequest struct {
	IAMRoleARN string `validate:"required"`
	InstallID  string `validate:"required"`
	Region     string `validate:"required"`

	TwoStepConfig *assumerole.TwoStepConfig `validate:"required"`
}

type DeleteEFSMountTargetsResponse struct{}

func (a *Activities) DeleteEFSMountTargets(ctx context.Context, req DeleteEFSMountTargetsRequest) (*DeleteEFSMountTargetsResponse, error) {
	efsClient, err := a.getEFSClient(ctx, req.IAMRoleARN, req.Region, req.TwoStepConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to get efs service: %w", err)
	}

	fs, err := a.getEFS(ctx, efsClient, req.InstallID)
	nfe := &efstypes.FileSystemNotFound{}
	if errors.As(err, &nfe) {
		return &DeleteEFSMountTargetsResponse{}, nil
	}

	mountTargets, err := a.getEFSMountTargets(ctx, efsClient, *fs.FileSystemId)
	if err != nil {
		return nil, fmt.Errorf("unable to get efs mount targets: %w", err)
	}

	for _, mountTargetID := range mountTargets {
		if err := a.deleteEFSMountTarget(ctx, efsClient, mountTargetID); err != nil {
			return nil, fmt.Errorf("unable to delete efs mount target: %w", err)
		}
	}

	return &DeleteEFSMountTargetsResponse{}, nil
}

func (a *Activities) getEFSMountTargets(ctx context.Context, efsClient *efs.Client, fsID string) ([]string, error) {
	resp, err := efsClient.DescribeMountTargets(ctx, &efs.DescribeMountTargetsInput{
		FileSystemId: generics.ToPtr(fsID),
	})
	nfe := &efstypes.FileSystemNotFound{}
	if errors.As(err, &nfe) {
		return nil, nil
	}

	mountTargetIDs := make([]string, 0)
	for _, mountTarget := range resp.MountTargets {
		mountTargetIDs = append(mountTargetIDs, *mountTarget.MountTargetId)
	}
	return mountTargetIDs, nil
}

func (a *Activities) deleteEFSMountTarget(ctx context.Context, efsClient *efs.Client, mountTargetID string) error {
	_, err := efsClient.DeleteMountTarget(ctx, &efs.DeleteMountTargetInput{
		MountTargetId: generics.ToPtr(mountTargetID),
	})
	nfe := &efstypes.MountTargetNotFound{}
	if errors.As(err, &nfe) {
		return nil
	}

	return nil
}

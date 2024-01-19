package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type CreateEFSMountTargetsRequest struct {
	IAMRoleARN string `validate:"required"`
	FsID       string
	Region     string `validate:"required"`

	VPCID           string
	SubnetIDs       []string
	SecurityGroupID string
}

type CreateEFSMountTargetsResponse struct{}

func (a *Activities) CreateEFSMountTargets(ctx context.Context, req CreateEFSMountTargetsRequest) (*CreateEFSMountTargetsResponse, error) {
	efsClient, err := a.getEFSClient(ctx, req.IAMRoleARN, req.Region)
	if err != nil {
		return nil, fmt.Errorf("unable to create efs client: %w", err)
	}

	for _, subnetID := range req.SubnetIDs {
		if _, err := efsClient.CreateMountTarget(ctx, &efs.CreateMountTargetInput{
			FileSystemId: generics.ToPtr(req.FsID),
			SubnetId:     generics.ToPtr(subnetID),
			SecurityGroups: []string{
				req.SecurityGroupID,
			},
		}); err != nil {
			alreadyExistsErr := &efstypes.MountTargetConflict{}
			if errors.As(err, &alreadyExistsErr); err != nil {
				return &CreateEFSMountTargetsResponse{}, nil
			}

			return nil, fmt.Errorf("unable to create mount target for for subnet %s: %w", subnetID, err)
		}
	}

	return &CreateEFSMountTargetsResponse{}, nil
}

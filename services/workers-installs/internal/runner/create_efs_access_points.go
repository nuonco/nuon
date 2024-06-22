package runner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type CreateEFSAccessPointsRequest struct {
	InstallID string
	FsID      string
	Region    string `validate:"required"`

	VPCID           string
	SubnetIDs       []string
	SecurityGroupID string

	Auth *credentials.Config `validate:"required"`
}

type CreateEFSAccessPointsResponse struct {
	AccessPointIDs []string
}

func (a *Activities) CreateEFSAccessPoints(ctx context.Context, req *CreateEFSAccessPointsRequest) (*CreateEFSAccessPointsResponse, error) {
	efsClient, err := a.getEFSClient(ctx, req.Region, req.Auth)
	if err != nil {
		return nil, fmt.Errorf("unable to create efs client: %w", err)
	}

	uid := generics.ToPtr(int64(100))
	gid := generics.ToPtr(int64(1000))

	accessPointIDs := make([]string, 0)
	for _, subnetID := range req.SubnetIDs {
		accessPoint, err := efsClient.CreateAccessPoint(ctx, &efs.CreateAccessPointInput{
			FileSystemId: generics.ToPtr(req.FsID),
			PosixUser: &efstypes.PosixUser{
				Uid: uid,
				Gid: gid,
			},
			RootDirectory: &efstypes.RootDirectory{
				CreationInfo: &efstypes.CreationInfo{
					OwnerUid:    uid,
					OwnerGid:    gid,
					Permissions: generics.ToPtr("755"),
				},
				Path: generics.ToPtr("/waypointserverdata"),
			},
			Tags: []efstypes.Tag{
				{
					Key:   generics.ToPtr("Name"),
					Value: generics.ToPtr(req.InstallID),
				},
			},
		})
		if err != nil {
			return nil, fmt.Errorf("unable to create mount target for for subnet %s: %w", subnetID, err)
		}

		accessPointIDs = append(accessPointIDs, *accessPoint.AccessPointId)
	}

	return &CreateEFSAccessPointsResponse{
		AccessPointIDs: accessPointIDs,
	}, nil
}

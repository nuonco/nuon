package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
)

type DeleteEFSAccessPointsRequest struct {
	InstallID string `validate:"required"`
	Region    string `validate:"required"`

	Auth *credentials.Config `validate:"required"`
}

type DeleteEFSAccessPointsResponse struct{}

func (a *Activities) DeleteEFSAccessPoints(ctx context.Context, req *DeleteEFSAccessPointsRequest) (*DeleteEFSAccessPointsResponse, error) {
	efsClient, err := a.getEFSClient(ctx, req.Region, req.Auth)
	if err != nil {
		return nil, fmt.Errorf("unable to get efs service: %w", err)
	}

	fs, err := a.getEFS(ctx, efsClient, req.InstallID)
	nfe := &efstypes.FileSystemNotFound{}
	if errors.As(err, &nfe) {
		return &DeleteEFSAccessPointsResponse{}, nil
	}

	accessPoints, err := a.getEFSAccessPoints(ctx, efsClient, *fs.FileSystemId)
	if err != nil {
		return nil, fmt.Errorf("unable to get efs mount targets: %w", err)
	}

	for _, accessPointID := range accessPoints {
		if err := a.deleteEFSAccessPoint(ctx, efsClient, accessPointID); err != nil {
			return nil, fmt.Errorf("unable to delete efs mount target: %w", err)
		}
	}

	return &DeleteEFSAccessPointsResponse{}, nil
}

func (a *Activities) getEFSAccessPoints(ctx context.Context, efsClient *efs.Client, fsID string) ([]string, error) {
	resp, err := efsClient.DescribeAccessPoints(ctx, &efs.DescribeAccessPointsInput{
		FileSystemId: generics.ToPtr(fsID),
	})
	nfe := &efstypes.FileSystemNotFound{}
	if errors.As(err, &nfe) {
		return nil, nil
	}

	accessPointIDs := make([]string, 0)
	for _, accessPoint := range resp.AccessPoints {
		accessPointIDs = append(accessPointIDs, *accessPoint.AccessPointId)
	}
	return accessPointIDs, nil
}

func (a *Activities) deleteEFSAccessPoint(ctx context.Context, efsClient *efs.Client, accessPointID string) error {
	_, err := efsClient.DeleteAccessPoint(ctx, &efs.DeleteAccessPointInput{
		AccessPointId: generics.ToPtr(accessPointID),
	})
	nfe := &efstypes.AccessPointNotFound{}
	if errors.As(err, &nfe) {
		return nil
	}

	return nil
}

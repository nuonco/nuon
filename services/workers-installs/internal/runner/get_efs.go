package runner

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	efstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (a *Activities) getEFS(ctx context.Context, efsClient *efs.Client, installID string) (*efstypes.FileSystemDescription, error) {
	fss, err := efsClient.DescribeFileSystems(ctx, &efs.DescribeFileSystemsInput{
		CreationToken: generics.ToPtr(installID),
	})

	nfe := &efstypes.FileSystemNotFound{}
	if errors.As(err, &nfe) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get efs: %w", err)
	}

	if len(fss.FileSystems) != 1 {
		return nil, fmt.Errorf("unable to find efs for install: %w", err)
	}

	return &fss.FileSystems[0], nil
}

package executors

import (
	"context"
	"fmt"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/s3downloader"
	"google.golang.org/protobuf/proto"
)

func (r *repo) GetDeploymentsPlan(ctx context.Context, key string) (*planv1.Plan, error) {
	client, err := s3downloader.New(r.DeploymentsBucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	byts, err := client.GetBlob(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	plan := &planv1.Plan{}
	if err := proto.Unmarshal(byts, plan); err != nil {
		return nil, fmt.Errorf("invalid plan: %w", err)
	}

	return plan, nil
}

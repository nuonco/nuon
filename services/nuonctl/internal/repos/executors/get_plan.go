package executors

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/nuonctl/internal/s3downloader"
	planv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/executors/v1/plan/v1"
	"google.golang.org/protobuf/proto"
)

func (r *repo) GetPlan(ctx context.Context, ref *planv1.PlanRef) (*planv1.Plan, error) {
	client, err := s3downloader.New(ref.Bucket,
		s3downloader.WithAssumeRoleARN(r.IAMRoleARN),
		s3downloader.WithAssumeRoleSessionName(assumeRoleSessionName))
	if err != nil {
		return nil, fmt.Errorf("unable to get downloader: %w", err)
	}

	byts, err := client.GetBlob(ctx, ref.BucketKey)
	if err != nil {
		return nil, fmt.Errorf("unable to get blob: %w", err)
	}

	plan := &planv1.Plan{}
	if err := proto.Unmarshal(byts, plan); err != nil {
		return nil, fmt.Errorf("invalid plan: %w", err)
	}

	return plan, nil
}

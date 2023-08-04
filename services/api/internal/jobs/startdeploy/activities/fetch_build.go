package activities

import (
	"context"
	"io"

	s3fetch "github.com/powertoolsdev/mono/pkg/deprecated/fetch/s3"
	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"google.golang.org/protobuf/proto"
)

func (a *activities) FetchBuildPlanJob(ctx context.Context, planref *planv1.PlanRef) (*planv1.Plan, error) {
	fetcher, err := s3fetch.New(
		a.v,
		s3fetch.WithBucketKey(planref.BucketKey),
		s3fetch.WithBucketName(planref.Bucket),
		s3fetch.WithRoleARN(planref.BucketAssumeRoleArn),
		s3fetch.WithRoleSessionName("deploy-fetch-build-plan"),
	)
	if err != nil {
		return nil, err
	}

	iorc, err := fetcher.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	defer iorc.Close()

	bs, err := io.ReadAll(iorc)
	if err != nil {
		return nil, err
	}

	var bp planv1.Plan
	if err = proto.Unmarshal(bs, &bp); err != nil {
		return nil, err
	}
	return &bp, nil
}

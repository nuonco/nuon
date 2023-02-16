package plan

import (
	"context"
	"fmt"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"google.golang.org/protobuf/proto"
)

type planUploader interface {
	uploadPlan(context.Context, s3BlobUploader, *planv1.PlanRef, *planv1.Plan) error
}

type planUploaderImpl struct{}

var _ planUploader = (*planUploaderImpl)(nil)

func (p *planUploaderImpl) uploadPlan(
	ctx context.Context,
	uploader s3BlobUploader,
	planRef *planv1.PlanRef,
	plan *planv1.Plan,
) error {
	byts, err := proto.Marshal(plan)
	if err != nil {
		return fmt.Errorf("unable to serialize plan: %w", err)
	}

	if err := uploader.UploadBlob(ctx, byts, planRef.BucketKey); err != nil {
		return fmt.Errorf("unable to upload plan: %w", err)
	}

	return nil
}

type s3BlobUploader interface {
	UploadBlob(context.Context, []byte, string) error
}

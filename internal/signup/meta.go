package signup

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/go-common/shortid"
	orgsv1 "github.com/powertoolsdev/protos/workflows/generated/types/orgs/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
	"github.com/powertoolsdev/workers-orgs/internal/meta"
	"go.temporal.io/sdk/workflow"
)

func getS3PrefixFromRequest(req *orgsv1.SignupRequest) (string, error) {
	shortID, err := shortid.ParseString(req.OrgId)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("org=%s", shortID), nil
}

func (a *Activities) FinishSignupRequest(ctx context.Context, req meta.FinishRequest) (meta.FinishResponse, error) {
	act := meta.NewFinishActivity()
	return act.FinishRequest(ctx, req)
}

// NOTE(jm): this is named StartSignup so we don't take up "StartRequest" in the deployments namespace
func (a *Activities) StartSignupRequest(ctx context.Context, req meta.StartRequest) (meta.StartResponse, error) {
	act := meta.NewStartActivity()
	return act.StartRequest(ctx, req)
}

func (w *wkflow) startWorkflow(ctx workflow.Context, req *orgsv1.SignupRequest) error {
	info := workflow.GetInfo(ctx)

	prefix, err := getS3PrefixFromRequest(req)
	if err != nil {
		return err
	}

	startReq := meta.StartRequest{
		MetadataBucket:              w.cfg.OrgsBucketName,
		MetadataBucketAssumeRoleARN: w.cfg.OrgsBucketAccessRoleArn,
		MetadataBucketPrefix:        prefix,
		Request:                     req,
		WorkflowInfo: meta.WorkflowInfo{
			ID: info.WorkflowExecution.ID,
		},
	}

	act := NewActivities(nil)
	if _, err = execStart(ctx, act, startReq); err != nil {
		return fmt.Errorf("unable to start workflow: %w", err)
	}

	return nil
}

// finishWorkflow calls the finish step
func (w *wkflow) finishWorkflow(ctx workflow.Context, req *orgsv1.SignupRequest, resp *orgsv1.SignupResponse, workflowErr error) {
	var err error
	defer func() {
		if err == nil {
			return
		}

		l := workflow.GetLogger(ctx)
		l.Debug("unable to finish workflow: %w", err)
	}()

	status := sharedv1.ResponseStatus_RESPONSE_STATUS_OK
	errMessage := ""
	if workflowErr != nil {
		status = sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR
		errMessage = workflowErr.Error()
	}

	prefix, err := getS3PrefixFromRequest(req)
	if err != nil {
		err = fmt.Errorf("unable to get s3 prefix: %w", err)
		return
	}

	finishReq := meta.FinishRequest{
		MetadataBucket:              w.cfg.OrgsBucketName,
		MetadataBucketAssumeRoleARN: w.cfg.OrgsBucketAccessRoleArn,
		MetadataBucketPrefix:        prefix,
		Response:                    resp,
		ResponseStatus:              status,
		ErrorMessage:                errMessage,
	}

	// exec activity
	act := NewActivities(nil)
	_, err = execFinish(ctx, act, finishReq)
	if err != nil {
		err = fmt.Errorf("unable to execute finish activity: %w", err)
	}
}

func execStart(
	ctx workflow.Context,
	act *Activities,
	req meta.StartRequest,
) (meta.StartResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := meta.StartResponse{}

	l.Debug("executing start activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.StartSignupRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

func execFinish(
	ctx workflow.Context,
	act *Activities,
	req meta.FinishRequest,
) (meta.FinishResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := meta.FinishResponse{}

	l.Debug("executing finish activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.FinishSignupRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

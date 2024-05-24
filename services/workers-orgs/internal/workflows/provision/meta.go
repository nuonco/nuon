package provision

import (
	"context"
	"fmt"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	meta "github.com/powertoolsdev/mono/pkg/workflows/meta"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	"go.temporal.io/sdk/workflow"
)

func (w *wkflow) startWorkflow(ctx workflow.Context, req *orgsv1.ProvisionRequest) error {
	info := workflow.GetInfo(ctx)
	prefix := prefix.OrgPath(req.OrgId)

	startReq := &sharedv1.StartActivityRequest{
		MetadataBucket:              w.cfg.OrgsBucketName,
		MetadataBucketAssumeRoleArn: w.cfg.OrgsBucketAccessRoleArn,
		MetadataBucketPrefix:        prefix,
		RequestRef:                  metaRequestFromReq(req),
		WorkflowInfo: &sharedv1.WorkflowInfo{
			Id: info.WorkflowExecution.ID,
		},
	}

	act := NewActivities()
	if _, err := execStart(ctx, act, startReq); err != nil {
		return fmt.Errorf("unable to start workflow: %w", err)
	}

	return nil
}

func metaRequestFromReq(req *orgsv1.ProvisionRequest) *sharedv1.RequestRef {
	return &sharedv1.RequestRef{
		Request: &sharedv1.RequestRef_OrgSignup{
			OrgSignup: req,
		},
	}
}

func metaResponseFromResponse(resp *orgsv1.ProvisionResponse) *sharedv1.ResponseRef {
	return &sharedv1.ResponseRef{
		Response: &sharedv1.ResponseRef_OrgSignup{
			OrgSignup: resp,
		},
	}
}

// NOTE(jm): none of the methods below this file should be modified.
func (a *Activities) FinishSignupRequest(ctx context.Context, req *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
	act := meta.NewFinishActivity()
	return act.FinishRequest(ctx, req)
}

func (a *Activities) StartSignupRequest(ctx context.Context, req *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
	act := meta.NewStartActivity()
	return act.StartRequest(ctx, req)
}

// finishWorkflow calls the finish step
func (w *wkflow) finishWorkflow(ctx workflow.Context, req *orgsv1.ProvisionRequest, resp *orgsv1.ProvisionResponse, workflowErr error) {
	var err error
	defer func() {
		if err == nil {
			return
		}

		l := workflow.GetLogger(ctx)
		l.Debug("unable to finish workflow: %w", err)
	}()

	prefix := prefix.OrgPath(req.OrgId)

	status := sharedv1.ResponseStatus_RESPONSE_STATUS_OK
	errMessage := ""
	if workflowErr != nil {
		status = sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR
		errMessage = workflowErr.Error()
	}

	finishReq := &sharedv1.FinishActivityRequest{
		MetadataBucket:              w.cfg.OrgsBucketName,
		MetadataBucketAssumeRoleArn: w.cfg.OrgsBucketAccessRoleArn,
		MetadataBucketPrefix:        prefix,
		ResponseRef:                 metaResponseFromResponse(resp),
		Status:                      status,
		ErrorMessage:                errMessage,
	}

	// exec activity
	act := NewActivities()
	_, err = execFinish(ctx, act, finishReq)
	if err != nil {
		err = fmt.Errorf("unable to execute finish activity: %w", err)
	}
}

func execStart(
	ctx workflow.Context,
	act *Activities,
	req *sharedv1.StartActivityRequest,
) (*sharedv1.StartActivityResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := &sharedv1.StartActivityResponse{}

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
	req *sharedv1.FinishActivityRequest,
) (*sharedv1.FinishActivityResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := &sharedv1.FinishActivityResponse{}

	l.Debug("executing finish activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.FinishSignupRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

package start

import (
	"context"
	"fmt"

	meta "github.com/powertoolsdev/mono/pkg/workflows-meta"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/deployments/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/shared/v1"
	"go.temporal.io/sdk/workflow"
)

func (w *wkflow) startWorkflow(ctx workflow.Context, req *deploymentsv1.StartRequest) error {
	info := workflow.GetInfo(ctx)
	prefix := getS3PrefixFromRequest(req)

	startReq := &sharedv1.StartActivityRequest{
		MetadataBucket:              w.cfg.DeploymentsBucket,
		MetadataBucketAssumeRoleArn: fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, req.OrgId),
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

func metaRequestFromReq(req *deploymentsv1.StartRequest) *sharedv1.RequestRef {
	return &sharedv1.RequestRef{
		Request: &sharedv1.RequestRef_DeploymentStart{
			DeploymentStart: req,
		},
	}
}

func metaResponseFromResponse(resp *deploymentsv1.StartResponse) *sharedv1.ResponseRef {
	return &sharedv1.ResponseRef{
		Response: &sharedv1.ResponseRef_DeploymentStart{
			DeploymentStart: resp,
		},
	}
}

// NOTE(jm): none of the methods below this file should be modified.
func (a *Activities) FinishStartRequest(ctx context.Context, req *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
	act := meta.NewFinishActivity()
	return act.FinishRequest(ctx, req)
}

func (a *Activities) StartStartRequest(ctx context.Context, req *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
	act := meta.NewStartActivity()
	return act.StartRequest(ctx, req)
}

// finishWorkflow calls the finish step
func (w *wkflow) finishWorkflow(ctx workflow.Context, req *deploymentsv1.StartRequest, resp *deploymentsv1.StartResponse, workflowErr error) {
	var err error
	defer func() {
		if err == nil {
			return
		}

		l := workflow.GetLogger(ctx)
		l.Debug("unable to finish workflow: %w", err)
	}()

	prefix := getS3PrefixFromRequest(req)

	status := sharedv1.ResponseStatus_RESPONSE_STATUS_OK
	errMessage := ""
	if workflowErr != nil {
		status = sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR
		errMessage = workflowErr.Error()
	}

	finishReq := &sharedv1.FinishActivityRequest{
		MetadataBucket:              w.cfg.DeploymentsBucket,
		MetadataBucketAssumeRoleArn: fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, req.OrgId),
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
	fut := workflow.ExecuteActivity(ctx, act.StartStartRequest, req)
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
	fut := workflow.ExecuteActivity(ctx, act.FinishStartRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

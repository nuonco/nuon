package provision

import (
	"context"
	"fmt"

	meta "github.com/powertoolsdev/mono/pkg/workflows-meta"
	"github.com/powertoolsdev/mono/pkg/workflows-meta/prefix"
	appsv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/apps/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/protos/workflows/generated/types/shared/v1"
	"go.temporal.io/sdk/workflow"
)

func (a *Activities) FinishProvisionRequest(ctx context.Context, req *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
	act := meta.NewFinishActivity()
	return act.FinishRequest(ctx, req)
}

func (a *Activities) StartProvisionRequest(ctx context.Context, req *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
	act := meta.NewStartActivity()
	return act.StartRequest(ctx, req)
}

func (w *Workflow) startWorkflow(ctx workflow.Context, req *appsv1.ProvisionRequest) error {
	info := workflow.GetInfo(ctx)

	startReq := &sharedv1.StartActivityRequest{
		MetadataBucket:              w.cfg.OrgsBucketName,
		MetadataBucketAssumeRoleArn: fmt.Sprintf(w.cfg.OrgsRoleTemplate, req.OrgId),
		MetadataBucketPrefix:        prefix.AppPath(req.OrgId, req.AppId),
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

func metaRequestFromReq(req *appsv1.ProvisionRequest) *sharedv1.RequestRef {
	return &sharedv1.RequestRef{
		Request: &sharedv1.RequestRef_AppProvision{
			AppProvision: req,
		},
	}
}

func metaResponseFromResponse(resp *appsv1.ProvisionResponse) *sharedv1.ResponseRef {
	return &sharedv1.ResponseRef{
		Response: &sharedv1.ResponseRef_AppsProvision{
			AppsProvision: resp,
		},
	}
}

// finishWorkflow calls the finish step
func (w *Workflow) finishWorkflow(ctx workflow.Context, req *appsv1.ProvisionRequest, resp *appsv1.ProvisionResponse, workflowErr error) {
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

	finishReq := &sharedv1.FinishActivityRequest{
		MetadataBucket:              w.cfg.OrgsBucketName,
		MetadataBucketAssumeRoleArn: fmt.Sprintf(w.cfg.OrgsRoleTemplate, req.OrgId),
		MetadataBucketPrefix:        prefix.AppPath(req.OrgId, req.AppId),
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
	fut := workflow.ExecuteActivity(ctx, act.StartProvisionRequest, req)
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
	fut := workflow.ExecuteActivity(ctx, act.FinishProvisionRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

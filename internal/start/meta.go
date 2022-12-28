package start

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/go-common/shortid"
	deploymentsv1 "github.com/powertoolsdev/protos/workflows/generated/types/deployments/v1"
	sharedv1 "github.com/powertoolsdev/protos/workflows/generated/types/shared/v1"
	"github.com/powertoolsdev/workers-deployments/internal/meta"
	"go.temporal.io/sdk/workflow"
)

func (a *Activities) FinishStartRequest(ctx context.Context, req meta.FinishRequest) (meta.FinishResponse, error) {
	act := meta.NewFinishActivity()
	return act.FinishRequest(ctx, req)
}

// NOTE(jm): this is named StartStart so we don't take up "StartRequest" in the deployments namespace
func (a *Activities) StartStartRequest(ctx context.Context, req meta.StartRequest) (meta.StartResponse, error) {
	act := meta.NewStartActivity()
	return act.StartRequest(ctx, req)
}

func (w *wkflow) startWorkflow(ctx workflow.Context, req *deploymentsv1.StartRequest) error {
	info := workflow.GetInfo(ctx)

	prefix, err := getS3PrefixFromRequest(req)
	if err != nil {
		return err
	}
	orgID, err := shortid.ParseString(req.OrgId)
	if err != nil {
		return fmt.Errorf("unable to parse org id into shortid: %w", err)
	}

	startReq := meta.StartRequest{
		MetadataBucket:              w.cfg.DeploymentsBucket,
		MetadataBucketAssumeRoleARN: fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, orgID),
		MetadataBucketPrefix:        prefix,
		Request:                     req,
		WorkflowInfo: meta.WorkflowInfo{
			ID: info.WorkflowExecution.ID,
		},
	}

	act := NewActivities()
	if _, err = execStart(ctx, act, startReq); err != nil {
		return fmt.Errorf("unable to start workflow: %w", err)
	}

	return nil
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

	orgID, err := shortid.ParseString(req.OrgId)
	if err != nil {
		err = fmt.Errorf("unable to parse orgID: %w", err)
		return
	}

	finishReq := meta.FinishRequest{
		MetadataBucket:              w.cfg.DeploymentsBucket,
		MetadataBucketAssumeRoleARN: fmt.Sprintf(w.cfg.OrgsDeploymentsRoleTemplate, orgID),
		MetadataBucketPrefix:        prefix,
		Response:                    resp,
		ResponseStatus:              status,
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
	req meta.StartRequest,
) (meta.StartResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := meta.StartResponse{}

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
	req meta.FinishRequest,
) (meta.FinishResponse, error) {
	l := workflow.GetLogger(ctx)
	resp := meta.FinishResponse{}

	l.Debug("executing finish activity", "request", req)
	fut := workflow.ExecuteActivity(ctx, act.FinishStartRequest, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

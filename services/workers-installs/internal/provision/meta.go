package provision

import (
	"context"
	"fmt"

	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	sharedv1 "github.com/powertoolsdev/mono/pkg/types/workflows/shared/v1"
	meta "github.com/powertoolsdev/mono/pkg/workflows/meta"
	"github.com/powertoolsdev/mono/pkg/workflows/meta/prefix"
	"go.temporal.io/sdk/workflow"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

func (w *wkflow) startWorkflow(ctx workflow.Context, req *installsv1.ProvisionRequest) error {
	info := workflow.GetInfo(ctx)
	prefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)

	startReq := &sharedv1.StartActivityRequest{
		MetadataBucket:              w.cfg.InstallationsBucket,
		MetadataBucketAssumeRoleArn: fmt.Sprintf(w.cfg.OrgInstallationsRoleTemplate, req.OrgId),
		MetadataBucketPrefix:        prefix,
		RequestRef:                  metaRequestFromReq(req),
		WorkflowInfo: &sharedv1.WorkflowInfo{
			Id: info.WorkflowExecution.ID,
		},
	}

	act := NewActivities(nil, nil, nil)
	if _, err := execStart(ctx, act, startReq); err != nil {
		return fmt.Errorf("unable to start workflow: %w", err)
	}

	return nil
}

func metaRequestFromReq(req *installsv1.ProvisionRequest) *sharedv1.RequestRef {
	return &sharedv1.RequestRef{
		Request: &sharedv1.RequestRef_InstallProvision{
			InstallProvision: req,
		},
	}
}

func metaResponseFromResponse(resp *installsv1.ProvisionResponse) *sharedv1.ResponseRef {
	return &sharedv1.ResponseRef{
		Response: &sharedv1.ResponseRef_InstallProvision{
			InstallProvision: resp,
		},
	}
}

// NOTE(jm): the following start and response activities need to properly emit notifications, once the meta workflows
// package works with them.
func (a *Activities) FinishProvisionRequest(ctx context.Context, req *sharedv1.FinishActivityRequest) (*sharedv1.FinishActivityResponse, error) {
	var err error

	var wkflowReq installsv1.ProvisionRequest
	if err = req.Other.UnmarshalTo(&wkflowReq); err != nil {
		return nil, fmt.Errorf("request was not set on input")
	}

	if req.Status != sharedv1.ResponseStatus_RESPONSE_STATUS_OK {
		err = a.notifier.sendErrorNotification(ctx, req.MetadataBucket, &wkflowReq, req.ErrorMessage)
	} else {
		err = a.notifier.sendSuccessNotification(ctx, req.MetadataBucket, &wkflowReq)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to send notification: %w", err)
	}

	act := meta.NewFinishActivity()
	return act.FinishRequest(ctx, req)
}

func (a *Activities) StartProvisionRequest(ctx context.Context, req *sharedv1.StartActivityRequest) (*sharedv1.StartActivityResponse, error) {
	if err := a.notifier.sendStartNotification(ctx, req.MetadataBucket, req.RequestRef.GetInstallProvision()); err != nil {
		return nil, fmt.Errorf("unable to send start notification: %w", err)
	}

	act := meta.NewStartActivity()
	return act.StartRequest(ctx, req)
}

// finishWorkflow calls the finish step
func (w *wkflow) finishWorkflow(ctx workflow.Context, req *installsv1.ProvisionRequest, resp *installsv1.ProvisionResponse, workflowErr error) {
	var err error
	defer func() {
		if err == nil {
			return
		}

		l := workflow.GetLogger(ctx)
		l.Debug("unable to finish workflow: %w", err)
	}()

	prefix := prefix.InstallPath(req.OrgId, req.AppId, req.InstallId)

	status := sharedv1.ResponseStatus_RESPONSE_STATUS_OK
	errMessage := ""
	if workflowErr != nil {
		status = sharedv1.ResponseStatus_RESPONSE_STATUS_ERROR
		errMessage = workflowErr.Error()
	}

	other, err := anypb.New(req)
	if err != nil {
		err = fmt.Errorf("unable to create any object: %w", err)
	}
	finishReq := &sharedv1.FinishActivityRequest{
		MetadataBucket:              w.cfg.InstallationsBucket,
		MetadataBucketAssumeRoleArn: fmt.Sprintf(w.cfg.OrgInstallationsRoleTemplate, req.OrgId),
		MetadataBucketPrefix:        prefix,
		ResponseRef:                 metaResponseFromResponse(resp),
		Status:                      status,
		ErrorMessage:                errMessage,
		Other:                       other,
	}

	// exec activity
	act := NewActivities(nil, nil, nil)
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

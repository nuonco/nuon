package iam

import (
	"fmt"
	"strings"

	iamv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/iam/v1"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/roles"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/structpb"
)

// DeprovisionIAM is a workflow that deprovisions all IAM for an org
func (w wkflow) DeprovisionIAM(ctx workflow.Context, req *iamv1.DeprovisionIAMRequest) (*iamv1.DeprovisionIAMResponse, error) {
	resp := &iamv1.DeprovisionIAMResponse{}
	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 2,
		},
	}
	act := NewActivities()
	ctx = workflow.WithActivityOptions(ctx, activityOpts)

	status := make(map[string]interface{})
	nameFns := map[string]func(string) string{
		"deployments":   roles.DeploymentsIAMName,
		"installations": roles.InstallationsIAMName,
		"installer":     roles.InstallerIAMName,
		"instances":     roles.InstancesIAMName,
		"odr":           roles.OdrIAMName,
		"orgs":          roles.OrgsIAMName,
		"secrets":       roles.SecretsIAMName,
	}
	for step, nameFn := range nameFns {
		if err := w.execDeprovisionRole(ctx,
			act,
			req,
			nameFn); err != nil {

			status[step] = fmt.Errorf("unable to delete IAM role: %w", err).Error()
			continue
		}
		status[step] = "ok"
	}

	respStruct, err := structpb.NewStruct(status)
	if err != nil {
		return resp, fmt.Errorf("unable to convert struct to proto: %w", err)
	}
	resp.Status = respStruct

	return resp, nil
}

func firstError(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *wkflow) execDeprovisionRole(ctx workflow.Context,
	act *Activities,
	req *iamv1.DeprovisionIAMRequest,
	nameFn func(string) string) error {

	arnPrefix := strings.Replace(w.cfg.OrgsAccountRootARN, ":root", "", 1)
	policyARN := fmt.Sprintf("%s:policy/orgs/%s/%s", arnPrefix, req.OrgId, nameFn(req.OrgId))

	deleteAttachmentReq := DeleteIAMRolePolicyAttachmentRequest{
		AssumeRoleARN: w.cfg.OrgsIAMAccessRoleArn,
		PolicyArn:     policyARN,
		RoleName:      nameFn(req.OrgId),
	}
	attachmentErr := execDeleteIAMRolePolicyAttachment(ctx, act, deleteAttachmentReq)

	deleteRoleReq := DeleteIAMRoleRequest{
		AssumeRoleARN: w.cfg.OrgsIAMAccessRoleArn,
		RoleName:      nameFn(req.OrgId),
	}
	_, roleErr := execDeleteIAMRole(ctx, act, deleteRoleReq)

	deletePolicyReq := DeleteIAMPolicyRequest{
		AssumeRoleARN: w.cfg.OrgsIAMAccessRoleArn,
		PolicyARN:     policyARN,
	}
	_, policyErr := execDeleteIAMPolicy(ctx, act, deletePolicyReq)

	return firstError(attachmentErr, roleErr, policyErr)
}

func execDeleteIAMPolicy(
	ctx workflow.Context,
	act *Activities,
	req DeleteIAMPolicyRequest,
) (DeleteIAMPolicyResponse, error) {
	var resp DeleteIAMPolicyResponse

	l := workflow.GetLogger(ctx)

	l.Debug("executing delete iam policy activity")
	fut := workflow.ExecuteActivity(ctx, act.DeleteIAMPolicy, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execDeleteIAMRole(
	ctx workflow.Context,
	act *Activities,
	req DeleteIAMRoleRequest,
) (DeleteIAMRoleResponse, error) {
	var resp DeleteIAMRoleResponse
	l := workflow.GetLogger(ctx)

	l.Debug("executing delete iam role activity")
	fut := workflow.ExecuteActivity(ctx, act.DeleteIAMRole, req)

	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execDeleteIAMRolePolicyAttachment(
	ctx workflow.Context,
	act *Activities,
	req DeleteIAMRolePolicyAttachmentRequest,
) error {
	l := workflow.GetLogger(ctx)

	l.Debug("executing delete iam role policy attachment activity")
	fut := workflow.ExecuteActivity(ctx, act.DeleteIAMRolePolicyAttachment, req)

	var resp DeleteIAMRolePolicyAttachmentResponse
	if err := fut.Get(ctx, &resp); err != nil {
		return err
	}

	return nil
}

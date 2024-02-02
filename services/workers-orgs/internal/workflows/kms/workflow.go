package kms

import (
	"fmt"
	"time"

	kmsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1/kms/v1"
	workers "github.com/powertoolsdev/mono/services/workers-orgs/internal"
	"github.com/powertoolsdev/mono/services/workers-orgs/internal/roles"
	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"
)

const (
	defaultActivityTimeout time.Duration = time.Second * 10
)

// NewWorkflow returns a new workflow executor
func NewWorkflow(cfg workers.Config) wkflow {
	return wkflow{
		cfg: cfg,
	}
}

type wkflow struct {
	cfg workers.Config
}

// ProvisionIAM is a workflow that creates org specific IAM roles in the designated orgs IAM account
func (w wkflow) ProvisionKMS(ctx workflow.Context, req *kmsv1.ProvisionKMSRequest) (*kmsv1.ProvisionKMSResponse, error) {
	l := log.With(workflow.GetLogger(ctx))
	act := NewActivities()
	activityOpts := workflow.ActivityOptions{
		ScheduleToCloseTimeout: defaultActivityTimeout,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOpts)
	resp := &kmsv1.ProvisionKMSResponse{}
	if err := req.Validate(); err != nil {
		return resp, fmt.Errorf("unable to validate request: %w", err)
	}

	if req.Reprovision {
		l.Info("reprovisioning, so assuming that kms key exists")
		getKeyReq := &GetKMSKeyRequest{
			AssumeRoleARN: w.cfg.OrgsKMSAccessRoleArn,
			KeyARN:        fmt.Sprintf("alias/org-%s", req.OrgId),
		}
		keyResp, err := execGetKMSKey(ctx, act, getKeyReq)
		if err == nil {
			resp.KmsKeyArn = keyResp.KeyArn
			resp.KmsKeyId = keyResp.KeyID
			return resp, nil
		}

		// if an error exists, attempt to re-run the workflow, regardless.
	}

	ckkReq := CreateKMSKeyRequest{
		AssumeRoleARN: w.cfg.OrgsKMSAccessRoleArn,
		KeyTags:       append(roles.DefaultTags(req.OrgId), [2]string{"Name", "org/" + req.OrgId}),
	}
	l.Debug("creating KMS key")
	ckkResp, err := execCreateKMSKey(ctx, act, ckkReq)
	if err != nil {
		return resp, fmt.Errorf("unable to create kms key: %w", err)
	}
	l.Debug("finished creating kms key", "key", ckkResp)
	resp.KmsKeyArn = ckkResp.KeyArn
	resp.KmsKeyId = ckkResp.KeyID

	l.Debug("creating KMS key policy")
	policy, err := roles.SecretsKMSKeyPolicy(req.SecretsIamRoleArn, w.cfg.OrgsKMSAccessRoleArn, w.cfg.OrgsAccountRootARN)
	if err != nil {
		return resp, fmt.Errorf("unable to get kms key policy: %w", err)
	}
	ckkpReq := CreateKMSKeyPolicyRequest{
		AssumeRoleARN: w.cfg.OrgsKMSAccessRoleArn,
		KeyID:         ckkResp.KeyID,
		Policy:        string(policy),
	}
	err = execCreateKMSKeyPolicy(ctx, act, ckkpReq)
	if err != nil {
		return resp, fmt.Errorf("unable to create kms key: %w", err)
	}

	l.Debug("creating KMS key alias")
	ckkaReq := CreateKMSKeyAliasRequest{
		AssumeRoleARN: w.cfg.OrgsKMSAccessRoleArn,
		KeyID:         ckkResp.KeyID,
		Alias:         fmt.Sprintf("alias/org-%s", req.OrgId),
	}
	err = execCreateKMSKeyAlias(ctx, act, ckkaReq)
	if err != nil {
		return resp, fmt.Errorf("unable to create kms alias: %w", err)
	}

	l.Debug("finished creating kms key", "key", ckkResp)
	return resp, nil
}

func execCreateKMSKey(
	ctx workflow.Context,
	act *Activities,
	req CreateKMSKeyRequest,
) (CreateKMSKeyResponse, error) {
	var resp CreateKMSKeyResponse

	l := workflow.GetLogger(ctx)

	l.Debug("executing create kms key activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateKMSKey, req)
	if err := fut.Get(ctx, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func execCreateKMSKeyPolicy(
	ctx workflow.Context,
	act *Activities,
	req CreateKMSKeyPolicyRequest,
) error {
	l := workflow.GetLogger(ctx)

	l.Debug("executing create kms key policy activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateKMSKeyPolicy, req)

	var resp CreateKMSKeyPolicyResponse
	if err := fut.Get(ctx, &resp); err != nil {
		return err
	}

	return nil
}

func execCreateKMSKeyAlias(
	ctx workflow.Context,
	act *Activities,
	req CreateKMSKeyAliasRequest,
) error {
	l := workflow.GetLogger(ctx)

	l.Debug("executing create kms key alias activity")
	fut := workflow.ExecuteActivity(ctx, act.CreateKMSKeyAlias, req)

	var resp CreateKMSKeyAliasResponse
	if err := fut.Get(ctx, &resp); err != nil {
		return err
	}

	return nil
}

func execGetKMSKey(
	ctx workflow.Context,
	act *Activities,
	req *GetKMSKeyRequest,
) (*GetKMSKeyResponse, error) {
	l := workflow.GetLogger(ctx)

	l.Debug("executing get kms key activity")
	fut := workflow.ExecuteActivity(ctx, act.GetKMSKey, req)

	var resp GetKMSKeyResponse
	if err := fut.Get(ctx, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

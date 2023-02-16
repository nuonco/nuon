package sandbox

import (
	"context"
	"fmt"

	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// TODO(jdt):
// - env vars?
// - un-hardcode install_role_arn

const (
	defaultTerraformVersion = "v1.3.9"
	defaultStateFilename    = "state.tf"
)

func (p *planner) Plan(ctx context.Context) (*planv1.Plan, error) {
	vars, err := p.vars()
	if err != nil {
		return nil, err
	}

	// NOTE(jdt): we should not try to default a version anywhere after this point
	vers := defaultTerraformVersion
	if p.sandbox.TerraformVersion != nil {
		vers = *p.sandbox.TerraformVersion
	}

	moduleKey := fmt.Sprintf("sandboxes/%s_%s.tar.gz", p.sandbox.SandboxSettings.Name, p.sandbox.SandboxSettings.Version)
	backendKey := fmt.Sprintf("%s/%s", p.Prefix(), defaultStateFilename)

	plan := &planv1.TerraformPlan{
		Id: p.sandbox.InstallId,
		Module: &planv1.Object{
			Bucket:            p.Request.Module.Name,
			Region:            p.Request.Module.Region,
			Key:               moduleKey,
			AssumeRoleDetails: p.Request.Module.AssumeRoleDetails,
		},
		Backend: &planv1.Object{
			Bucket: p.Request.Backend.Name,
			Region: p.Request.Backend.Region,
			Key:    backendKey,
		},
		Vars:             vars,
		RunType:          p.sandbox.RunType,
		TerraformVersion: vers,
		Outputs:          map[string]*planv1.Object{},
	}

	return &planv1.Plan{Actual: &planv1.Plan_TerraformPlan{TerraformPlan: plan}}, nil
}

func (p *planner) vars() (*structpb.Struct, error) {
	installID := p.sandbox.InstallId
	sboxSettings := p.sandbox.SandboxSettings

	awsSettings, ok := p.sandbox.AccountSettings.(*planv1.Sandbox_Aws)
	if !ok {
		return nil, fmt.Errorf("unsupported account settings")
	}

	return structpb.NewStruct(map[string]interface{}{
		"nuon_id":                           installID,
		"region":                            awsSettings.Aws.Region,
		"assume_role_arn":                   awsSettings.Aws.RoleArn,
		"install_role_arn":                  "arn:aws:iam::618886478608:role/install-k8s-admin-stage",
		"waypoint_odr_namespace":            installID,
		"waypoint_odr_service_account_name": fmt.Sprintf("waypoint-odr-%s", installID),
		"tags": map[string]string{
			"nuon_sandbox_name":    sboxSettings.Name,
			"nuon_sandbox_version": sboxSettings.Version,
			"nuon_install_id":      installID,
			"nuon_app_id":          p.sandbox.AppId,
		},
	})
}

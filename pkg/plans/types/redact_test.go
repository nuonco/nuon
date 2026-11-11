package plantypes

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	assumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

func planWithSecrets() *CompositePlan {
	return &CompositePlan{
		BuildPlan: &BuildPlan{
			Src: &GitSource{URL: "https://acme:ghs_secret@github.com/acme/app.git", Ref: "main"},
			Dst: &configs.OCIRegistryRepository{
				Repository: "123456789012.dkr.ecr.us-west-2.amazonaws.com/acme",
				OCIAuth:    &configs.OCIRegistryAuth{Username: "AWS", Password: "ecr-secret"},
				ECRAuth: &awscredentials.Config{Static: &awscredentials.StaticCredentials{
					AccessKeyID:     "AKIAEXAMPLE",
					SecretAccessKey: "aws-secret",
					SessionToken:    "aws-session",
				}},
			},
			ContainerImagePullPlan: &ContainerImagePullPlan{
				RepoCfg: &configs.OCIRegistryRepository{
					OCIAuth: &configs.OCIRegistryAuth{Username: "oauth2accesstoken", Password: "gar-secret"},
				},
			},
		},
	}
}

func actionPlanWithSecrets() *CompositePlan {
	return &CompositePlan{
		ActionWorkflowRunPlan: &ActionWorkflowRunPlan{
			InstallID: "inst_example",
			BuiltinEnvVars: map[string]string{
				"NUON_ORG_ID":    "org_example",
				"NUON_API_URL":   "https://api.example.com",
				"NUON_API_TOKEN": "runner-token",
			},
			Steps: []*ActionWorkflowRunStepPlan{{
				InterpolatedEnvVars: map[string]string{
					"NUON_API_TOKEN": "copied-token",
					"LOG_LEVEL":      "info",
				},
			}},
			AWSAuth: &awscredentials.Config{
				AssumeRole: &awscredentials.AssumeRoleConfig{
					RoleARN: "arn:aws:iam::123456789012:role/example",
					TwoStepConfig: &assumerole.TwoStepConfig{
						SrcIAMRoleARN: "arn:aws:iam::123456789012:role/jump",
						SrcStaticCredentials: assumerole.StaticCredentials{
							AccessKeyID:     "AKIAEXAMPLE",
							SecretAccessKey: "src-secret",
							SessionToken:    "src-session",
						},
					},
				},
			},
		},
	}
}

func assertActionPlanRedacted(t *testing.T, plan *ActionWorkflowRunPlan) {
	t.Helper()

	assert.Equal(t, "org_example", plan.BuiltinEnvVars["NUON_ORG_ID"])
	assert.Equal(t, "https://api.example.com", plan.BuiltinEnvVars["NUON_API_URL"])
	assert.Empty(t, plan.BuiltinEnvVars["NUON_API_TOKEN"])
	assert.Empty(t, plan.Steps[0].InterpolatedEnvVars["NUON_API_TOKEN"])
	assert.Equal(t, "info", plan.Steps[0].InterpolatedEnvVars["LOG_LEVEL"])

	src := plan.AWSAuth.AssumeRole.TwoStepConfig.SrcStaticCredentials
	assert.Equal(t, "arn:aws:iam::123456789012:role/example", plan.AWSAuth.AssumeRole.RoleARN)
	assert.Equal(t, "arn:aws:iam::123456789012:role/jump", plan.AWSAuth.AssumeRole.TwoStepConfig.SrcIAMRoleARN)
	assert.Equal(t, "AKIAEXAMPLE", src.AccessKeyID)
	assert.Empty(t, src.SecretAccessKey)
	assert.Empty(t, src.SessionToken)
}

func assertRedacted(t *testing.T, bp *BuildPlan) {
	t.Helper()

	assert.Equal(t, "https://github.com/acme/app.git", bp.Src.URL)
	assert.Equal(t, "main", bp.Src.Ref)
	assert.Equal(t, "AWS", bp.Dst.OCIAuth.Username)
	assert.Empty(t, bp.Dst.OCIAuth.Password)
	assert.Equal(t, "AKIAEXAMPLE", bp.Dst.ECRAuth.Static.AccessKeyID)
	assert.Empty(t, bp.Dst.ECRAuth.Static.SecretAccessKey)
	assert.Empty(t, bp.Dst.ECRAuth.Static.SessionToken)
	assert.Equal(t, "123456789012.dkr.ecr.us-west-2.amazonaws.com/acme", bp.Dst.Repository)
	assert.Equal(t, "oauth2accesstoken", bp.ContainerImagePullPlan.RepoCfg.OCIAuth.Username)
	assert.Empty(t, bp.ContainerImagePullPlan.RepoCfg.OCIAuth.Password)
}

func TestCompositePlanRedactCredentials(t *testing.T) {
	tests := map[string]struct {
		plan  *CompositePlan
		check func(t *testing.T, cp *CompositePlan)
	}{
		"blanks secrets and keeps identifiers": {
			plan: planWithSecrets(),
			check: func(t *testing.T, cp *CompositePlan) {
				assertRedacted(t, cp.BuildPlan)
			},
		},
		"leaves a url without credentials alone": {
			plan: &CompositePlan{SandboxRunPlan: &SandboxRunPlan{
				GitSource: &GitSource{URL: "https://github.com/acme/app.git"},
			}},
			check: func(t *testing.T, cp *CompositePlan) {
				assert.Equal(t, "https://github.com/acme/app.git", cp.SandboxRunPlan.GitSource.URL)
			},
		},
		"blanks the action run token": {
			plan: actionPlanWithSecrets(),
			check: func(t *testing.T, cp *CompositePlan) {
				assertActionPlanRedacted(t, cp.ActionWorkflowRunPlan)
			},
		},
		"empty plan": {
			plan: &CompositePlan{},
			check: func(t *testing.T, cp *CompositePlan) {
				assert.True(t, cp.IsEmpty())
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			tc.plan.RedactCredentials()
			tc.check(t, tc.plan)
		})
	}

	var nilPlan *CompositePlan
	assert.NotPanics(t, nilPlan.RedactCredentials)
}

func TestRedactPlanJSON(t *testing.T) {
	withSecrets, err := json.Marshal(planWithSecrets().BuildPlan)
	require.NoError(t, err)
	actionPlan, err := json.Marshal(actionPlanWithSecrets().ActionWorkflowRunPlan)
	require.NoError(t, err)

	tests := map[string]struct {
		input string
		check func(t *testing.T, out string)
	}{
		"blanks secrets and keeps identifiers": {
			input: string(withSecrets),
			check: func(t *testing.T, out string) {
				var bp BuildPlan
				require.NoError(t, json.Unmarshal([]byte(out), &bp))
				assertRedacted(t, &bp)
			},
		},
		"blanks the action run token": {
			input: string(actionPlan),
			check: func(t *testing.T, out string) {
				var plan ActionWorkflowRunPlan
				require.NoError(t, json.Unmarshal([]byte(out), &plan))
				assertActionPlanRedacted(t, &plan)
			},
		},
		"returns input unchanged when nothing to redact": {
			input: `{"container_image_pull_plan": {"image_url": "example.com/app:tag"}}`,
			check: func(t *testing.T, out string) {
				assert.Equal(t, `{"container_image_pull_plan": {"image_url": "example.com/app:tag"}}`, out)
			},
		},
		"returns input unchanged when secret fields are already empty": {
			input: `{"dst_registry": {"OCIAuth": {"Username": "AWS", "Password": ""}}}`,
			check: func(t *testing.T, out string) {
				assert.Equal(t, `{"dst_registry": {"OCIAuth": {"Username": "AWS", "Password": ""}}}`, out)
			},
		},
		"empty input": {
			input: "",
			check: func(t *testing.T, out string) {
				assert.Empty(t, out)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			out, err := RedactPlanJSON(tc.input)
			require.NoError(t, err)
			tc.check(t, out)
		})
	}

	_, err = RedactPlanJSON("{not json")
	assert.Error(t, err)
}

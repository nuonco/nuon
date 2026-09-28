package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	apiPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/api"
)

type mcpStackSetup struct {
	Headline              string         `json:"headline"`
	Status                string         `json:"status,omitempty"`
	StatusDescription     string         `json:"status_description,omitempty"`
	LastChecked           string         `json:"last_checked,omitempty"`
	Cloud                 string         `json:"cloud"`
	StackName             string         `json:"stack_name,omitempty"`
	Region                string         `json:"region,omitempty"`
	TemplateURL           string         `json:"template_url,omitempty"`
	QuickLinkURL          string         `json:"quick_link_url,omitempty"`
	CreateCommand         string         `json:"create_command,omitempty"`
	UpdateCommand         string         `json:"update_command,omitempty"`
	ConsoleURL            string         `json:"console_url,omitempty"`
	InputsTfvars          string         `json:"inputs_tfvars,omitempty"`
	ProviderTfvars        string         `json:"provider_tfvars,omitempty"`
	SecretsTfvars         string         `json:"secrets_tfvars,omitempty"`
	TerraformCloneCommand string         `json:"terraform_clone_command,omitempty"`
	TerraformBackend      string         `json:"terraform_backend,omitempty"`
	TerraformApplyCommand string         `json:"terraform_apply_command,omitempty"`
	TFModuleSource        string         `json:"tf_module_source,omitempty"`
	TFModuleVersion       string         `json:"tf_module_version,omitempty"`
	Outputs               map[string]any `json:"outputs,omitempty"`
}

func (s *service) attachStackSetups(ctx context.Context, orgID string, steps []app.WorkflowStep, summaries []mcpWorkflowStepSummary) {
	if len(steps) != len(summaries) {
		return
	}
	for i, step := range steps {
		if step.Name != app.AwaitInstallStackStepName || step.StepTargetID == "" {
			continue
		}
		setup := s.mcpStackSetup(ctx, orgID, step.StepTargetID)
		if setup == nil {
			continue
		}
		summaries[i].StackSetup = setup
	}
}

func (s *service) mcpStackSetup(ctx context.Context, orgID, versionID string) *mcpStackSetup {
	var version app.InstallStackVersion
	err := s.db.WithContext(ctx).
		Select("id", "install_id", "status", "template_url", "quick_link_url", "stack_name", "terraform_contents").
		Where("id = ? AND org_id = ?", versionID, orgID).
		First(&version).Error
	if err != nil {
		return nil
	}

	var install app.Install
	_ = s.db.WithContext(ctx).
		Select("id").
		Preload("AWSAccount").
		Preload("AzureAccount").
		Where("id = ? AND org_id = ?", version.InstallID, orgID).
		First(&install).Error

	cloud := stackSetupCloud(install, version.QuickLinkURL)
	region := ""
	if install.AWSAccount != nil {
		region = install.AWSAccount.Region
	}
	if region == "" {
		region = queryParam(version.QuickLinkURL, "region")
	}

	stackName := version.StackName
	if stackName == "" {
		stackName = fragmentParam(version.QuickLinkURL, "stackName")
	}
	if stackName == "" && version.InstallID != "" {
		stackName = "nuon-" + version.InstallID
	}

	headline := "Install stack is waiting to run"
	if version.Status.Status == app.InstallStackVersionStatusActive {
		headline = "Install stack up and running"
	}

	setup := &mcpStackSetup{
		Headline:          headline,
		Status:            string(version.Status.Status),
		StatusDescription: version.Status.StatusHumanDescription,
		Cloud:             cloud,
		StackName:         stackName,
		Region:            region,
		TemplateURL:       version.TemplateURL,
		QuickLinkURL:      version.QuickLinkURL,
		TFModuleSource:    "nuonco/stack/" + cloud,
		TFModuleVersion:   "~> 1.0",
	}

	if cloud == "aws" && version.TemplateURL != "" {
		regionForCmd := region
		if regionForCmd == "" {
			regionForCmd = "<YOUR_REGION>"
		}
		setup.CreateCommand = awsCloudFormationCommand("create-stack", stackName, version.TemplateURL, regionForCmd)
		setup.UpdateCommand = awsCloudFormationCommand("update-stack", stackName, version.TemplateURL, regionForCmd)
		if region != "" {
			setup.ConsoleURL = fmt.Sprintf(
				"https://console.aws.amazon.com/cloudformation/home?region=%s#/stacks/events?filteringText=%s&filteringStatus=active&viewNested=true",
				url.QueryEscape(region),
				url.QueryEscape(stackName),
			)
		} else if stackName != "" {
			setup.ConsoleURL = fmt.Sprintf(
				"https://console.aws.amazon.com/cloudformation/home#/stacks?filteringText=%s",
				url.QueryEscape(stackName),
			)
		}
	}

	inputs, provider, secrets := s.stackTfvars(ctx, version)
	if inputs != "" || secrets != "" {
		setup.InputsTfvars = inputs
		setup.ProviderTfvars = provider
		setup.SecretsTfvars = secrets
		if cloud == "aws" || cloud == "gcp" {
			setup.TerraformCloneCommand = fmt.Sprintf("git clone https://github.com/nuonco/install-stacks.git\ncd install-stacks/%s", cloud)
			setup.TerraformApplyCommand = "terraform init && terraform apply"
			setup.TerraformBackend = terraformBackend(cloud, version.InstallID)
		}
	}

	var run app.InstallStackVersionRun
	err = s.db.WithContext(ctx).
		Where("install_stack_version_id = ? AND org_id = ?", version.ID, orgID).
		Order("created_at DESC").
		First(&run).Error
	if err == nil {
		if !run.UpdatedAt.IsZero() {
			setup.LastChecked = apiPkg.MCPTime(run.UpdatedAt)
		}
		if len(run.DataContents) > 0 {
			setup.Outputs = run.DataContents
		}
	}

	return setup
}

func (s *service) stackTfvars(ctx context.Context, version app.InstallStackVersion) (inputs, provider, secrets string) {
	inputs, provider, secrets = parseTfvarsEnvelope(version.TerraformContents)
	if inputs != "" || secrets != "" {
		return inputs, provider, secrets
	}

	var contents struct {
		Contents []byte
	}
	err := s.db.WithContext(ctx).
		Model(&app.InstallStackVersion{}).
		Select("contents").
		Where("id = ?", version.ID).
		Take(&contents).Error
	if err != nil {
		return "", "", ""
	}
	return parseTfvarsEnvelope(contents.Contents)
}

func parseTfvarsEnvelope(body []byte) (inputs, provider, secrets string) {
	if len(body) == 0 {
		return "", "", ""
	}
	var envelope struct {
		InputsTFVars   string `json:"inputs_tfvars"`
		ProviderTFVars string `json:"provider_tfvars"`
		SecretsTFVars  string `json:"secrets_tfvars"`
		LegacyTFVars   string `json:"tfvars"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return "", "", ""
	}
	inputs = envelope.InputsTFVars
	if inputs == "" && envelope.SecretsTFVars == "" {
		inputs = envelope.LegacyTFVars
	}
	return inputs, envelope.ProviderTFVars, envelope.SecretsTFVars
}

func stackSetupCloud(install app.Install, quickLink string) string {
	switch {
	case install.AWSAccount != nil:
		return "aws"
	case install.AzureAccount != nil:
		return "azure"
	}
	switch {
	case strings.Contains(quickLink, "console.aws.amazon.com"):
		return "aws"
	case strings.Contains(quickLink, "portal.azure.com"):
		return "azure"
	default:
		return "gcp"
	}
}

func awsCloudFormationCommand(verb, stackName, templateURL, region string) string {
	if isS3TemplateURL(templateURL) {
		return fmt.Sprintf(
			"aws cloudformation %s \\\n  --stack-name %s \\\n  --template-url %s \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region %s",
			verb, stackName, templateURL, region,
		)
	}
	return fmt.Sprintf(
		"curl -sLo template.json %q \\\n  && aws cloudformation %s \\\n  --stack-name %s \\\n  --template-body file://template.json \\\n  --capabilities CAPABILITY_NAMED_IAM \\\n  --region %s",
		templateURL, verb, stackName, region,
	)
}

func isS3TemplateURL(templateURL string) bool {
	return strings.Contains(templateURL, "s3.amazonaws.com") || strings.Contains(templateURL, ".s3.")
}

func terraformBackend(cloud, installID string) string {
	if cloud == "gcp" {
		return fmt.Sprintf(`terraform {
  backend "gcs" {
    bucket = "<your-state-bucket>"
    prefix = "nuon/%s"
  }
}`, installID)
	}
	return fmt.Sprintf(`terraform {
  backend "s3" {
    bucket = "<your-state-bucket>"
    key    = "nuon/%s/terraform.tfstate"
    region = "<your-state-bucket-region>"
  }
}`, installID)
}

func queryParam(rawURL, key string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Query().Get(key)
}

func fragmentParam(rawURL, key string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	frag := u.Fragment
	idx := strings.Index(frag, "?")
	if idx < 0 {
		return ""
	}
	vals, err := url.ParseQuery(frag[idx+1:])
	if err != nil {
		return ""
	}
	return vals.Get(key)
}

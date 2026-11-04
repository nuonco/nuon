package build

import (
	"fmt"
	"time"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type InstallConfigInput struct {
	ApprovalOption           app.InstallApprovalOption             `json:"approval_option"`
	Telemetry                *config.InstallTelemetry              `json:"telemetry,omitempty"`
	Labels                   map[string]string                     `json:"labels,omitempty"`
	VPCNestedTemplateURL     *string                               `json:"vpc_nested_template_url,omitempty"`
	RunnerNestedTemplateURL  *string                               `json:"runner_nested_template_url,omitempty"`
	CustomNestedStacks       []config.CustomNestedStack            `json:"custom_nested_stacks,omitempty"`
	AWSStackParameters       map[string]string                     `json:"-"`
	AWSNestedStackParameters map[string]config.CloudFormationStack `json:"-"`
	RunnerImageAutoUpdate    *bool                                 `json:"runner_image_auto_update,omitempty" extensions:"x-nullable"`
	RunnerImageID            *string                               `json:"runner_image_id,omitempty"`
	ActionTimeout            *time.Duration                        `json:"action_timeout,omitempty" swaggertype:"primitive,integer" extensions:"x-nullable"`
}

func InstallConfigInputFromConfig(install *config.Install) InstallConfigInput {
	in := InstallConfigInput{
		ApprovalOption: app.InstallApprovalOption(install.ApprovalOption),
		Telemetry:      install.Telemetry,
	}
	if overrides := install.StackOverrides; overrides != nil {
		if overrides.VPCNestedTemplateURL != "" {
			in.VPCNestedTemplateURL = generics.ToPtr(overrides.VPCNestedTemplateURL)
		}
		if overrides.RunnerNestedTemplateURL != "" {
			in.RunnerNestedTemplateURL = generics.ToPtr(overrides.RunnerNestedTemplateURL)
		}
		in.CustomNestedStacks = overrides.CustomNestedStacks
		in.RunnerImageAutoUpdate = overrides.RunnerImageAutoUpdate
		if overrides.RunnerImageID != "" {
			in.RunnerImageID = generics.ToPtr(overrides.RunnerImageID)
		}
	}
	if overrides := install.CloudFormation; overrides != nil {
		in.AWSStackParameters = overrides.Root.Parameters
		in.AWSNestedStackParameters = overrides.NestedStacks
	}
	if install.ActionTimeout != "" {
		timeout, err := config.ParseActionTimeout(install.ActionTimeout)
		if err == nil {
			in.ActionTimeout = &timeout
		}
	}
	return in
}

func (in InstallConfigInput) ValidateActionTimeout() error {
	if in.ActionTimeout == nil {
		return nil
	}
	if *in.ActionTimeout < config.MinActionTimeout || *in.ActionTimeout > config.MaxActionTimeout {
		return fmt.Errorf("action_timeout must be between %s and %s", config.MinActionTimeout, config.MaxActionTimeout)
	}
	return nil
}

func InstallConfig(in InstallConfigInput, installID string) *app.InstallConfig {
	obj := &app.InstallConfig{
		InstallID:                installID,
		ApprovalOption:           in.ApprovalOption,
		VPCNestedTemplateURL:     in.VPCNestedTemplateURL,
		RunnerNestedTemplateURL:  in.RunnerNestedTemplateURL,
		CustomNestedStacks:       in.CustomNestedStacks,
		AWSStackParameters:       in.AWSStackParameters,
		AWSNestedStackParameters: in.AWSNestedStackParameters,
		RunnerImageAutoUpdate:    in.RunnerImageAutoUpdate,
		RunnerImageID:            in.RunnerImageID,
		ActionTimeout:            in.ActionTimeout,
		Labeled:                  labels.Labeled{Labels: labels.Labels(in.Labels)},
	}
	if in.Telemetry != nil {
		obj.TelemetryEnabled = in.Telemetry.Enabled
	}
	return obj
}

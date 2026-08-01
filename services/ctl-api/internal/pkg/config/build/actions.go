package build

import (
	"time"

	"github.com/lib/pq"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type ActionWorkflowInput struct {
	AppID            string
	AppConfigID      string
	OrgID            string
	ActionWorkflowID string

	Timeout time.Duration

	DependencyIDs []string
	References    []string

	BreakGlassRole        string
	Role                  string
	EnableKubeConfig      *bool
	KubernetesContextName string
}

// ActionWorkflowConfig builds the row; the caller attaches triggers and steps.
func ActionWorkflowConfig(in ActionWorkflowInput) *app.ActionWorkflowConfig {
	// Written for action steps unless the config opts out.
	enableKubeConfig := true
	if in.EnableKubeConfig != nil {
		enableKubeConfig = *in.EnableKubeConfig
	}

	return &app.ActionWorkflowConfig{
		AppID:                  in.AppID,
		AppConfigID:            in.AppConfigID,
		OrgID:                  in.OrgID,
		ActionWorkflowID:       in.ActionWorkflowID,
		Timeout:                in.Timeout,
		ComponentDependencyIDs: pq.StringArray(in.DependencyIDs),
		References:             pq.StringArray(in.References),
		BreakGlassRoleARN:      generics.NewNullString(in.BreakGlassRole),
		Role:                   in.Role,
		EnableKubeConfig:       generics.NewNullBoolFromPtr(&enableKubeConfig),
		KubernetesContextName:  in.KubernetesContextName,
	}
}

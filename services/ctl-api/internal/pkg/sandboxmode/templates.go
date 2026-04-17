package sandboxmode

import (
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/sandboxmode/templates"
)

// SandboxLogTemplate represents a pre-built log template that can be used
// to populate sandbox config log lines.
type SandboxLogTemplate struct {
	Key         string   `json:"key"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Type        string   `json:"type"` // e.g. "failing-action", "kube-action", "success"
	Lines       []string `json:"lines"`
}

// SandboxPlanTemplate represents a pre-built plan template that can be used
// to populate sandbox config plan contents.
type SandboxPlanTemplate struct {
	Key         string `json:"key"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Type        string `json:"type"` // e.g. "noop", "s3", "database", "full-sandbox"
	Contents    string `json:"contents"`
}

// SandboxTemplates is the response for the templates endpoint.
type SandboxTemplates struct {
	LogTemplates         []SandboxLogTemplate  `json:"log_templates"`
	PlanTemplates        []SandboxPlanTemplate `json:"plan_templates"`
	PlanDisplayTemplates []SandboxPlanTemplate `json:"plan_display_templates"`
	StateTemplates       []SandboxPlanTemplate `json:"state_templates"`
}

func DefaultSandboxTemplates() SandboxTemplates {
	return SandboxTemplates{
		LogTemplates:         defaultLogTemplates(),
		PlanTemplates:        defaultPlanTemplates(),
		PlanDisplayTemplates: defaultPlanDisplayTemplates(),
		StateTemplates:       defaultStateTemplates(),
	}
}

func defaultLogTemplates() []SandboxLogTemplate {
	logTmpls := templates.LogTemplates()
	result := make([]SandboxLogTemplate, 0, len(logTmpls))
	for _, t := range logTmpls {
		logType := "success"
		if t.IsNoop {
			logType = "noop"
		}
		result = append(result, SandboxLogTemplate{
			Key:         t.Key,
			Description: t.Description,
			Category:    categoryFromJobTypes(t.JobTypes),
			Type:        logType,
			Lines:       strings.Split(t.Contents, "\n"),
		})
	}
	return result
}

func defaultPlanTemplates() []SandboxPlanTemplate {
	planTmpls := templates.PlanTemplates()
	result := make([]SandboxPlanTemplate, 0, len(planTmpls))
	for _, t := range planTmpls {
		planType := t.Key
		if t.IsNoop {
			planType = "noop"
		}
		result = append(result, SandboxPlanTemplate{
			Key:         t.Key,
			Description: t.Description,
			Category:    categoryFromKey(t.Key),
			Type:        planType,
			Contents:    t.Contents,
		})
	}
	return result
}

func defaultPlanDisplayTemplates() []SandboxPlanTemplate {
	displayTmpls := templates.PlanDisplayTemplates()
	result := make([]SandboxPlanTemplate, 0, len(displayTmpls))
	for _, t := range displayTmpls {
		planType := t.Key
		if t.IsNoop {
			planType = "noop"
		}
		result = append(result, SandboxPlanTemplate{
			Key:         t.Key,
			Description: t.Description,
			Category:    categoryFromKey(t.Key),
			Type:        planType,
			Contents:    t.Contents,
		})
	}
	return result
}

func defaultStateTemplates() []SandboxPlanTemplate {
	stateTmpls := templates.StateTemplates()
	result := make([]SandboxPlanTemplate, 0, len(stateTmpls))
	for _, t := range stateTmpls {
		stateType := t.Key
		if t.IsNoop {
			stateType = "noop"
		}
		result = append(result, SandboxPlanTemplate{
			Key:         t.Key,
			Description: t.Description,
			Category:    categoryFromKey(t.Key),
			Type:        stateType,
			Contents:    t.Contents,
		})
	}
	return result
}

// categoryFromJobTypes infers the legacy category from job types.
func categoryFromJobTypes(jobTypes []string) string {
	if len(jobTypes) == 0 {
		return "deploy"
	}
	jt := jobTypes[0]
	switch {
	case strings.Contains(jt, "build"):
		return "build"
	case strings.Contains(jt, "sync"):
		return "sync"
	case strings.Contains(jt, "actions"):
		return "actions"
	default:
		return "deploy"
	}
}

// categoryFromKey infers the legacy plan category from the template key.
func categoryFromKey(key string) string {
	switch {
	case strings.HasPrefix(key, "terraform"):
		return "terraform"
	case strings.HasPrefix(key, "helm"):
		return "helm"
	case strings.HasPrefix(key, "kube"):
		return "kubernetes"
	case strings.HasPrefix(key, "pulumi"):
		return "pulumi"
	default:
		return "other"
	}
}

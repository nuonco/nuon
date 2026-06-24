package deployerrors

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

// ComponentBuildUnavailableErrorType is the discriminator for a deploy that
// cannot proceed because the component it targets has no deployable build.
const ComponentBuildUnavailableErrorType compositeerrors.Type = "deploy.component_build_unavailable"

// ComponentBuildUnavailableReason describes why the build could not be deployed.
type ComponentBuildUnavailableReason string

const (
	// ComponentBuildUnavailableReasonFailed means the latest build for the
	// component is in a terminal failure state (error / policy_failed).
	ComponentBuildUnavailableReasonFailed ComponentBuildUnavailableReason = "failed"

	// ComponentBuildUnavailableReasonMissing means no build exists for the
	// component yet, so there is no artifact to deploy.
	ComponentBuildUnavailableReasonMissing ComponentBuildUnavailableReason = "missing"
)

// ComponentBuildUnavailableError is the typed payload for a deploy that cannot
// run because its component build is unavailable. It implements
// compositeerrors.CompositeError so it can be frozen onto the owning deploy row
// and rendered in the dashboard, guiding the user to (re)build the component.
type ComponentBuildUnavailableError struct {
	Reason ComponentBuildUnavailableReason `json:"reason"`

	ComponentID   string `json:"component_id,omitempty"`
	ComponentName string `json:"component_name,omitempty"`

	BuildID                string `json:"build_id,omitempty"`
	BuildStatus            string `json:"build_status,omitempty"`
	BuildStatusDescription string `json:"build_status_description,omitempty"`
}

var _ compositeerrors.CompositeError = (*ComponentBuildUnavailableError)(nil)

func (e *ComponentBuildUnavailableError) Error() string {
	name := e.ComponentName
	if name == "" {
		name = "component"
	}
	if e.Reason == ComponentBuildUnavailableReasonMissing {
		return fmt.Sprintf("No build found for %s", name)
	}
	return fmt.Sprintf("Build for %s failed", name)
}

func (e *ComponentBuildUnavailableError) Type() compositeerrors.Type {
	return ComponentBuildUnavailableErrorType
}

func (e *ComponentBuildUnavailableError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityError
}

func (e *ComponentBuildUnavailableError) Sections() []compositeerrors.Section {
	name := e.ComponentName
	if name == "" {
		name = "this component"
	}

	var why string
	if e.Reason == ComponentBuildUnavailableReasonMissing {
		why = fmt.Sprintf("Deploying %s needs a build, but it hasn't been built yet.", name)
	} else {
		why = fmt.Sprintf("Deploying %s needs a build that completed successfully. Its most recent build failed, so the deploy can't continue.", name)
	}
	sections := []compositeerrors.Section{
		{
			Heading: "Why",
			Body:    why,
		},
	}

	if e.BuildStatus != "" || e.BuildStatusDescription != "" {
		var body string
		if e.BuildID != "" {
			body += fmt.Sprintf("Build: `%s`\n\n", e.BuildID)
		}
		if e.BuildStatus != "" {
			body += fmt.Sprintf("Status: `%s`\n\n", e.BuildStatus)
		}
		if e.BuildStatusDescription != "" {
			body += "```\n" + e.BuildStatusDescription + "\n```"
		}
		sections = append(sections, compositeerrors.Section{
			Heading: "Build status",
			Body:    body,
		})
	}

	var fix string
	if e.Reason == ComponentBuildUnavailableReasonMissing {
		fix = fmt.Sprintf("Build %s, wait for the build to become active, then retry the deploy.", name)
	} else {
		fix = fmt.Sprintf("Fix what caused the build to fail, then rebuild %s with the latest config. Once the build is active, retry the deploy.", name)
	}
	sections = append(sections, compositeerrors.Section{
		Heading: "How to fix",
		Body:    fix,
	})

	return sections
}

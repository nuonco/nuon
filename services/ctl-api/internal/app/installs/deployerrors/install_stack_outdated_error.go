package deployerrors

import (
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

const InstallStackOutdatedErrorType compositeerrors.Type = "preflight.install_stack_outdated"

type InstallStackOutdatedError struct {
	InstallID string `json:"install_id"`
}

func (e *InstallStackOutdatedError) Error() string {
	return "Install stack is out of date"
}

func (e *InstallStackOutdatedError) Type() compositeerrors.Type {
	return InstallStackOutdatedErrorType
}

func (e *InstallStackOutdatedError) Severity() compositeerrors.Severity {
	return compositeerrors.SeverityWarning
}

func (e *InstallStackOutdatedError) Hints() compositeerrors.Hints {
	return compositeerrors.NewHints().WithDocsURL("https://docs.nuon.co/guides/reprovision-installs")
}

func (e *InstallStackOutdatedError) Sections() []compositeerrors.Section {
	return []compositeerrors.Section{
		compositeerrors.MarkdownSection(
			"Why this matters",
			"The install does not have an applied stack that matches the stack configuration required by this workflow. Workflows can fail or behave unexpectedly until the current stack is provisioned in the cloud account.",
		),
		compositeerrors.MarkdownSection(
			"Reprovision from the dashboard",
			"Open the install, select **Manage**, then select **Reprovision stack**. Review the operation role and component deployment options, then start the workflow. At the **await install stack** step, follow the generated instructions to apply the stack update in the cloud account.",
		),
		compositeerrors.CodeSection(
			"Reprovision with the CLI",
			fmt.Sprintf("nuon installs stacks reprovision --install-id %s", e.InstallID),
		),
	}
}
